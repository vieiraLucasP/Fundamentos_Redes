package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"ring-network/internal/queue"
	"ring-network/pkg/config"
	"ring-network/pkg/message"
)

type MachineStatus struct {
	MachineName      string
	HasToken         bool
	QueueSize        int
	LastActivity     time.Time
	TokensProcessed  int
	MessagesSent     int
	MessagesReceived int
	ErrorsDetected   int
	TokensGenerated  int
}

type Machine struct {
	config           *config.Config
	conn             *net.UDPConn
	queue            *queue.MessageQueue
	hasToken         bool
	running          bool
	mutex            sync.RWMutex
	lastActivity     time.Time
	status           *MachineStatus
	tokenTimeout     *time.Timer
	waitingForData   bool
	currentDataMsg   *message.DataMessage
	errorProbability float64
}

func NewMachine(cfg *config.Config) (*Machine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %v", err)
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("erro ao resolver endereço: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar socket UDP: %v", err)
	}

	machine := &Machine{
		config:           cfg,
		conn:             conn,
		queue:            queue.NewMessageQueue(10),
		hasToken:         false,
		running:          false,
		lastActivity:     time.Now(),
		waitingForData:   false,
		errorProbability: 0.1,
		status: &MachineStatus{
			MachineName:  cfg.MachineName,
			HasToken:     false,
			LastActivity: time.Now(),
		},
	}

	return machine, nil
}

func (m *Machine) Start() {
	m.mutex.Lock()
	m.running = true
	m.mutex.Unlock()

	log.Printf("[%s] Máquina iniciada na porta %d", m.config.MachineName, m.config.ListenPort)
	if m.config.GeneratesToken {
		go func() {
			time.Sleep(1 * time.Second)
			m.generateInitialToken()
		}()

		go m.tokenWatchdog()
	}

	buffer := make([]byte, 1024)
	for m.isRunning() {
		m.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := m.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("[%s] Erro ao ler dados: %v", m.config.MachineName, err)
			continue
		}

		data := string(buffer[:n])
		log.Printf("[%s] Recebido de %s: %s", m.config.MachineName, addr, data)

		m.handleReceivedData(data)
	}
}

func (m *Machine) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.running = false
	if m.conn != nil {
		m.conn.Close()
	}
	if m.tokenTimeout != nil {
		m.tokenTimeout.Stop()
	}

	log.Printf("[%s] Máquina parada", m.config.MachineName)
}

func (m *Machine) handleReceivedData(data string) {
	m.updateLastActivity()

	if message.IsTokenPacket(data) {
		m.handleToken()
		return
	}

	dataMsg, err := message.ParseDataPacket(data)
	if err != nil {
		log.Printf("[%s] Erro ao parsear pacote de dados: %v", m.config.MachineName, err)
		return
	}

	m.handleDataPacket(dataMsg)
}

func (m *Machine) handleToken() {
	log.Printf("[%s] Token recebido", m.config.MachineName)

	m.mutex.Lock()
	m.hasToken = true
	m.status.HasToken = true
	m.status.TokensProcessed++
	m.mutex.Unlock()

	if m.tokenTimeout != nil {
		m.tokenTimeout.Stop()
	}

	m.tokenTimeout = time.AfterFunc(time.Duration(m.config.TokenTime)*time.Second, func() {
		m.processToken()
	})
}

func (m *Machine) processToken() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.hasToken {
		return
	}

	if !m.queue.IsEmpty() {
		queuedMsg := m.queue.Peek()
		if queuedMsg != nil {
			dataMsg := message.CreateDataPacket(m.config.MachineName, queuedMsg.Destination, queuedMsg.Content)

			if queuedMsg.Destination == "TODOS" {
				log.Printf("[%s] Enviando mensagem BROADCAST: %s", m.config.MachineName, queuedMsg.Content)
			} else {
				if dataMsg.IntroduceError(m.errorProbability) {
					log.Printf("[%s] Erro introduzido na mensagem para %s", m.config.MachineName, queuedMsg.Destination)
				}
			}

			m.waitingForData = true
			m.currentDataMsg = dataMsg

			m.sendPacket(dataMsg.RawData)
			m.status.MessagesSent++

			log.Printf("[%s] Mensagem enviada para %s: %s", m.config.MachineName, queuedMsg.Destination, queuedMsg.Content)
		}
	} else {
		log.Printf("[%s] Fila vazia, passando token", m.config.MachineName)
		m.passToken()
	}
}

func (m *Machine) handleDataPacket(dataMsg *message.DataMessage) {
	log.Printf("[%s] Pacote de dados recebido: %s", m.config.MachineName, dataMsg.String())

	if dataMsg.Destination == m.config.MachineName || dataMsg.Destination == "TODOS" {
		m.handleMessageForThisMachine(dataMsg)
	} else if dataMsg.Origin == m.config.MachineName {
		m.handleReturnedMessage(dataMsg)
	} else {
		m.forwardMessage(dataMsg)
	}
}

func (m *Machine) handleMessageForThisMachine(dataMsg *message.DataMessage) {
	m.status.MessagesReceived++

	if dataMsg.Destination == "TODOS" {
		log.Printf("[%s] Mensagem BROADCAST recebida de %s: %s", m.config.MachineName, dataMsg.Origin, dataMsg.Message)

		if dataMsg.Origin == m.config.MachineName {
			m.mutex.Lock()
			m.waitingForData = false
			m.currentDataMsg = nil
			m.queue.RemoveFirstMessage()
			m.passToken()
			m.mutex.Unlock()
			return
		}

		m.forwardMessage(dataMsg)
		return
	}

	if dataMsg.VerifyIntegrity() {
		log.Printf("[%s] Mensagem recebida de %s: %s", m.config.MachineName, dataMsg.Origin, dataMsg.Message)
		dataMsg.SetControl(message.ControlACK)
	} else {
		log.Printf("[%s] Erro detectado na mensagem de %s", m.config.MachineName, dataMsg.Origin)
		dataMsg.SetControl(message.ControlNAK)
		m.status.ErrorsDetected++
	}

	m.sendPacket(dataMsg.RawData)
}

func (m *Machine) handleReturnedMessage(dataMsg *message.DataMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if dataMsg.Destination == "TODOS" {
		log.Printf("[%s] Mensagem BROADCAST completou o ciclo", m.config.MachineName)
		m.queue.RemoveFirstMessage()
		m.waitingForData = false
		m.currentDataMsg = nil
		m.passToken()
		return
	}

	if !m.waitingForData || m.currentDataMsg == nil {
		log.Printf("[%s] Mensagem retornada inesperada", m.config.MachineName)
		return
	}

	m.waitingForData = false
	m.currentDataMsg = nil

	switch dataMsg.Control {
	case message.ControlACK:
		log.Printf("[%s] ACK recebido para mensagem para %s", m.config.MachineName, dataMsg.Destination)
		m.queue.RemoveFirstMessage()

	case message.ControlNAK:
		log.Printf("[%s] NAK recebido para mensagem para %s - será retransmitida", m.config.MachineName, dataMsg.Destination)
		m.queue.IncrementRetries()

	case message.ControlMachineNotExists:
		log.Printf("[%s] Máquina %s não existe ou está desligada", m.config.MachineName, dataMsg.Destination)
		m.queue.RemoveFirstMessage()
	}

	m.passToken()
}

func (m *Machine) forwardMessage(dataMsg *message.DataMessage) {
	log.Printf("[%s] Repassando mensagem de %s para %s", m.config.MachineName, dataMsg.Origin, dataMsg.Destination)
	m.sendPacket(dataMsg.RawData)
}

func (m *Machine) passToken() {
	m.hasToken = false
	m.status.HasToken = false

	tokenPacket := message.CreateTokenPacket()
	m.sendPacket(tokenPacket)

	log.Printf("[%s] Token enviado para próxima máquina", m.config.MachineName)
}

func (m *Machine) sendPacket(data string) error {
	addr, err := net.ResolveUDPAddr("udp", m.config.NextMachineAddr)
	if err != nil {
		return fmt.Errorf("erro ao resolver endereço: %v", err)
	}

	_, err = m.conn.WriteToUDP([]byte(data), addr)
	if err != nil {
		return fmt.Errorf("erro ao enviar pacote: %v", err)
	}

	return nil
}

func (m *Machine) generateInitialToken() {
	log.Printf("[%s] Gerando token inicial", m.config.MachineName)

	m.mutex.Lock()
	m.status.TokensGenerated++
	m.mutex.Unlock()

	tokenPacket := message.CreateTokenPacket()
	err := m.sendPacket(tokenPacket)
	if err != nil {
		log.Printf("[%s] Erro ao enviar token inicial: %v", m.config.MachineName, err)
	}
}

func (m *Machine) QueueMessage(destination, content string) error {
	return m.queue.Enqueue(destination, content)
}

func (m *Machine) GetStatus() MachineStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := *m.status
	status.QueueSize = m.queue.Size()
	status.LastActivity = m.lastActivity

	return status
}

func (m *Machine) GetMessageQueue() []*message.QueuedMessage {
	return m.queue.GetAll()
}

func (m *Machine) GenerateToken() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.hasToken {
		return fmt.Errorf("máquina já possui o token")
	}

	go m.generateInitialToken()
	return nil
}

func (m *Machine) isRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

func (m *Machine) updateLastActivity() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lastActivity = time.Now()
}

func (m *Machine) tokenWatchdog() {
	if !m.config.GeneratesToken {
		return
	}

	maxTokenCirculationTime := time.Duration(m.config.TokenTime*3*2+3) * time.Second
	ticker := time.NewTicker(maxTokenCirculationTime)
	defer ticker.Stop()

	lastTokenSeen := time.Now()

	for {
		select {
		case <-ticker.C:
			m.mutex.RLock()
			hasToken := m.hasToken
			running := m.running
			timeSinceLastToken := time.Since(lastTokenSeen)
			m.mutex.RUnlock()

			if !running {
				return
			}

			if timeSinceLastToken > maxTokenCirculationTime && !hasToken {
				log.Printf("[%s] Token perdido! (último visto há %v) Gerando novo token...",
					m.config.MachineName, timeSinceLastToken)
				m.generateInitialToken()
				lastTokenSeen = time.Now()
			}

		default:
			m.mutex.RLock()
			if m.hasToken {
				lastTokenSeen = time.Now()
			}
			m.mutex.RUnlock()

			time.Sleep(1 * time.Second)
		}
	}
}
