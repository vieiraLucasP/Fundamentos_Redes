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

// MachineStatus representa o status da máquina
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

// Machine representa uma máquina na rede em anel
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
	errorProbability float64 // Probabilidade de inserir erro (0.0 a 1.0)
}

// NewMachine cria uma nova máquina
func NewMachine(cfg *config.Config) (*Machine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %v", err)
	}

	// Criar socket UDP para escutar
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
		queue:            queue.NewMessageQueue(10), // Máximo 10 mensagens
		hasToken:         false,
		running:          false,
		lastActivity:     time.Now(),
		waitingForData:   false,
		errorProbability: 0.1, // 10% de chance de erro
		status: &MachineStatus{
			MachineName:  cfg.MachineName,
			HasToken:     false,
			LastActivity: time.Now(),
		},
	}

	return machine, nil
}

// Start inicia a execução da máquina
func (m *Machine) Start() {
	m.mutex.Lock()
	m.running = true
	m.mutex.Unlock()

	log.Printf("[%s] Máquina iniciada na porta %d", m.config.MachineName, m.config.ListenPort)
	// Se esta máquina gera o token inicial, iniciar o processo
	if m.config.GeneratesToken {
		go func() {
			time.Sleep(1 * time.Second) // Aguardar inicialização básica
			m.generateInitialToken()
		}()

		// Configurar controle de token perdido
		go m.tokenWatchdog()
	}

	// Loop principal de recepção de pacotes
	buffer := make([]byte, 1024)
	for m.isRunning() {
		m.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := m.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout normal, continuar
			}
			log.Printf("[%s] Erro ao ler dados: %v", m.config.MachineName, err)
			continue
		}

		data := string(buffer[:n])
		
		// Verificar se é um token ou um pacote de dados
		if message.IsTokenPacket(data) {
			log.Printf("[%s] Recebido de %s: %s", m.config.MachineName, addr, data)
		} else {
			// Para pacotes de dados, não mostrar o conteúdo bruto
			log.Printf("[%s] Recebido pacote de dados de %s", m.config.MachineName, addr)
		}

		m.handleReceivedData(data)
	}
}

// Stop para a execução da máquina
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

// handleReceivedData processa dados recebidos
func (m *Machine) handleReceivedData(data string) {
	m.updateLastActivity()

	// Verificar se é um token
	if message.IsTokenPacket(data) {
		m.handleToken()
		return
	}

	// Verificar se é um pacote de dados
	dataMsg, err := message.ParseDataPacket(data)
	if err != nil {
		log.Printf("[%s] Erro ao parsear pacote de dados: %v", m.config.MachineName, err)
		return
	}

	m.handleDataPacket(dataMsg)
}

// handleToken processa o recebimento de um token
func (m *Machine) handleToken() {
	log.Printf("[%s] Token recebido", m.config.MachineName)

	m.mutex.Lock()
	m.hasToken = true
	m.status.HasToken = true
	m.status.TokensProcessed++
	m.mutex.Unlock()

	// Configurar timeout para o token
	if m.tokenTimeout != nil {
		m.tokenTimeout.Stop()
	}

	// Aguardar o tempo configurado antes de processar o token
	m.tokenTimeout = time.AfterFunc(time.Duration(m.config.TokenTime)*time.Second, func() {
		m.processToken()
	})
}

// processToken processa o token (envia mensagem ou passa o token)
func (m *Machine) processToken() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.hasToken {
		return // Token já foi processado
	}

	// Verificar se há mensagens na fila
	if !m.queue.IsEmpty() {
		// Pegar a primeira mensagem da fila
		queuedMsg := m.queue.Peek()
		if queuedMsg != nil {
			// Criar pacote de dados
			dataMsg := message.CreateDataPacket(m.config.MachineName, queuedMsg.Destination, queuedMsg.Content)

			// Caso especial para broadcast (TODOS)
			if queuedMsg.Destination == "TODOS" {
				log.Printf("[%s] Enviando mensagem BROADCAST: %s", m.config.MachineName, queuedMsg.Content)
				// Não aplicar módulo de falhas para broadcast
			} else {
				// Aplicar módulo de inserção de falhas apenas para unicast
				if dataMsg.IntroduceError(m.errorProbability) {
					log.Printf("[%s] Erro introduzido na mensagem para %s", m.config.MachineName, queuedMsg.Destination)
				}
			}

			// Marcar que estamos aguardando o retorno desta mensagem
			m.waitingForData = true
			m.currentDataMsg = dataMsg

			// Enviar mensagem
			m.sendPacket(dataMsg.RawData)
			m.status.MessagesSent++

			log.Printf("[%s] Mensagem enviada para %s: %s", m.config.MachineName, queuedMsg.Destination, queuedMsg.Content)
			// IMPORTANTE: Token só será passado quando dados retornarem (em handleReturnedMessage)
		}
	} else {
		// Fila vazia: passar o token imediatamente conforme especificação
		log.Printf("[%s] Fila vazia, passando token", m.config.MachineName)
		m.passToken()
	}
}

// handleDataPacket processa um pacote de dados recebido
func (m *Machine) handleDataPacket(dataMsg *message.DataMessage) {
	// Verificar se a mensagem é para esta máquina
	if dataMsg.Destination == m.config.MachineName {
		// Mensagem destinada a esta máquina - mostrar detalhes completos
		log.Printf("[%s] Pacote de dados recebido para mim: %s", m.config.MachineName, dataMsg.String())
		m.handleMessageForThisMachine(dataMsg)
	} else if dataMsg.Destination == "TODOS" {
		// Mensagem de broadcast - mostrar detalhes
		log.Printf("[%s] Pacote de dados broadcast recebido: %s", m.config.MachineName, dataMsg.String())
		m.handleMessageForThisMachine(dataMsg)
	} else if dataMsg.Origin == m.config.MachineName {
		// Mensagem de volta para o originador
		log.Printf("[%s] Pacote de dados retornado: %s", m.config.MachineName, dataMsg.String())
		m.handleReturnedMessage(dataMsg)
	} else {
		// Mensagem para outra máquina, repassar sem mostrar o conteúdo
		m.forwardMessage(dataMsg)
	}
}

// handleMessageForThisMachine processa mensagem destinada a esta máquina
func (m *Machine) handleMessageForThisMachine(dataMsg *message.DataMessage) {
	m.status.MessagesReceived++

	// Caso especial para broadcast (TODOS)
	if dataMsg.Destination == "TODOS" {
		// Para broadcast, apenas exibir a mensagem e manter o controle como maquinanaoexiste
		log.Printf("[%s] Mensagem BROADCAST recebida de %s: %s", m.config.MachineName, dataMsg.Origin, dataMsg.Message)
		
		// Se a origem for esta máquina, remover da fila e passar o token
		if dataMsg.Origin == m.config.MachineName {
			m.mutex.Lock()
			m.waitingForData = false
			m.currentDataMsg = nil
			m.queue.RemoveFirstMessage()
			m.passToken()
			m.mutex.Unlock()
			return
		}
		
		// Repassar a mensagem sem alterar o controle
		m.forwardMessage(dataMsg)
		return
	}

	// Verificar integridade da mensagem para mensagens unicast
	if dataMsg.VerifyIntegrity() {
		// Mensagem íntegra - mostrar que recebeu a mensagem sem expor o conteúdo
		log.Printf("[%s] Mensagem privada recebida de %s", m.config.MachineName, dataMsg.Origin)
		// Processar a mensagem internamente (sem mostrar no log)
		dataMsg.SetControl(message.ControlACK)
	} else {
		// Mensagem com erro
		log.Printf("[%s] Erro detectado na mensagem de %s", m.config.MachineName, dataMsg.Origin)
		dataMsg.SetControl(message.ControlNAK)
		m.status.ErrorsDetected++
	}

	// Enviar mensagem de volta
	m.sendPacket(dataMsg.RawData)
}

// handleReturnedMessage processa mensagem que voltou para o originador
func (m *Machine) handleReturnedMessage(dataMsg *message.DataMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Caso especial para broadcast (TODOS)
	if dataMsg.Destination == "TODOS" {
		log.Printf("[%s] Mensagem BROADCAST completou o ciclo", m.config.MachineName)
		// Remover mensagem da fila
		m.queue.RemoveFirstMessage()
		// Resetar estado de espera
		m.waitingForData = false
		m.currentDataMsg = nil
		// Passar o token para a próxima máquina
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
		// Remover mensagem da fila
		m.queue.RemoveFirstMessage()

	case message.ControlNAK:
		log.Printf("[%s] NAK recebido para mensagem para %s - será retransmitida", m.config.MachineName, dataMsg.Destination)
		// Incrementar tentativas, mas manter na fila
		m.queue.IncrementRetries()

	case message.ControlMachineNotExists:
		log.Printf("[%s] Máquina %s não existe ou está desligada", m.config.MachineName, dataMsg.Destination)
		// Remover mensagem da fila
		m.queue.RemoveFirstMessage()
	}

	// Passar o token para a próxima máquina
	m.passToken()
}

// forwardMessage repassa uma mensagem para a próxima máquina
func (m *Machine) forwardMessage(dataMsg *message.DataMessage) {
	// Não mostrar o conteúdo da mensagem ao repassar, apenas origem e destino
	log.Printf("[%s] Repassando mensagem de %s para %s (conteúdo privado)", 
		m.config.MachineName, dataMsg.Origin, dataMsg.Destination)
	m.sendPacket(dataMsg.RawData)
}

// passToken envia o token para a próxima máquina
func (m *Machine) passToken() {
	m.hasToken = false
	m.status.HasToken = false

	tokenPacket := message.CreateTokenPacket()
	m.sendPacket(tokenPacket)

	log.Printf("[%s] Token enviado para próxima máquina", m.config.MachineName)
}

// sendPacket envia um pacote para a próxima máquina
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

// generateInitialToken gera o token inicial
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

// QueueMessage adiciona uma mensagem à fila
func (m *Machine) QueueMessage(destination, content string) error {
	return m.queue.Enqueue(destination, content)
}

// GetStatus retorna o status atual da máquina
func (m *Machine) GetStatus() MachineStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := *m.status
	status.QueueSize = m.queue.Size()
	status.LastActivity = m.lastActivity

	return status
}

// GetMessageQueue retorna uma cópia da fila de mensagens
func (m *Machine) GetMessageQueue() []*message.QueuedMessage {
	return m.queue.GetAll()
}

// GenerateToken gera um novo token (para controle manual)
func (m *Machine) GenerateToken() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.hasToken {
		return fmt.Errorf("máquina já possui o token")
	}

	go m.generateInitialToken()
	return nil
}

// isRunning verifica se a máquina está em execução
func (m *Machine) isRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// updateLastActivity atualiza o timestamp da última atividade
func (m *Machine) updateLastActivity() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lastActivity = time.Now()
}

// tokenWatchdog monitora a circulação do token e regenera se perdido
func (m *Machine) tokenWatchdog() {
	if !m.config.GeneratesToken {
		return // Apenas a máquina geradora controla o token
	}

	// Tempo máximo para o token circular: (tempo_token * num_máquinas * 2) + margem
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

			// Se não vemos o token há muito tempo e não estamos com ele
			if timeSinceLastToken > maxTokenCirculationTime && !hasToken {
				log.Printf("[%s] ⚠️  Token perdido! (último visto há %v) Gerando novo token...",
					m.config.MachineName, timeSinceLastToken)
				m.generateInitialToken()
				lastTokenSeen = time.Now()
			}

		default:
			// Atualizar timestamp quando vemos o token
			m.mutex.RLock()
			if m.hasToken {
				lastTokenSeen = time.Now()
			}
			m.mutex.RUnlock()

			time.Sleep(1 * time.Second)
		}
	}
}
