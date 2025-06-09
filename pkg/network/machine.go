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

// MachineStatus armazena informações sobre o estado atual da máquina
// Usado para monitoramento e exibição de estatísticas
type MachineStatus struct {
	MachineName      string    // Nome da máquina
	HasToken         bool      // Indica se a máquina possui o token atualmente
	QueueSize        int       // Número de mensagens na fila
	LastActivity     time.Time // Timestamp da última atividade
	TokensProcessed  int       // Número de tokens processados
	MessagesSent     int       // Número de mensagens enviadas
	MessagesReceived int       // Número de mensagens recebidas
	ErrorsDetected   int       // Número de erros de CRC detectados
	TokensGenerated  int       // Número de tokens gerados por esta máquina
}

// Machine representa uma máquina na rede em anel
// Implementa a lógica de processamento de mensagens e token
type Machine struct {
	config           *config.Config       // Configuração da máquina
	conn             *net.UDPConn         // Conexão UDP para comunicação
	queue            *queue.MessageQueue  // Fila de mensagens para envio
	hasToken         bool                 // Indica se possui o token
	running          bool                 // Indica se a máquina está em execução
	mutex            sync.RWMutex         // Mutex para acesso concorrente
	lastActivity     time.Time            // Timestamp da última atividade
	status           *MachineStatus       // Status atual da máquina
	tokenTimeout     *time.Timer          // Timer para processamento do token
	waitingForData   bool                 // Indica se está aguardando resposta
	currentDataMsg   *message.DataMessage // Mensagem atual sendo processada
	errorProbability float64              // Probabilidade de introduzir erro
}

// NewMachine cria uma nova instância de máquina com a configuração fornecida
// Inicializa a conexão UDP e as estruturas internas
func NewMachine(cfg *config.Config) (*Machine, error) {
	// Valida a configuração antes de prosseguir
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %v", err)
	}

	// Configura o endereço UDP para escuta
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("erro ao resolver endereço: %v", err)
	}

	// Cria o socket UDP para comunicação
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar socket UDP: %v", err)
	}

	// Inicializa a máquina com valores padrão
	machine := &Machine{
		config:           cfg,
		conn:             conn,
		queue:            queue.NewMessageQueue(10), // Fila com capacidade para 10 mensagens
		hasToken:         false,
		running:          false,
		lastActivity:     time.Now(),
		waitingForData:   false,
		errorProbability: 0.1, // 10% de chance de introduzir erro
		status: &MachineStatus{
			MachineName:  cfg.MachineName,
			HasToken:     false,
			LastActivity: time.Now(),
		},
	}

	return machine, nil
}

// Start inicia a operação da máquina
// Executa o loop principal de recebimento de pacotes
func (m *Machine) Start() {
	m.mutex.Lock()
	m.running = true
	m.mutex.Unlock()

	log.Printf("[%s] Máquina iniciada na porta %d", m.config.MachineName, m.config.ListenPort)

	// Se esta máquina é responsável por gerar o token inicial
	if m.config.GeneratesToken {
		// Gera o token inicial após um pequeno delay
		go func() {
			time.Sleep(1 * time.Second)
			m.generateInitialToken()
		}()

		// Inicia o watchdog para monitorar a circulação do token
		go m.tokenWatchdog()
	}

	// Loop principal de recebimento de pacotes
	buffer := make([]byte, 1024)
	for m.isRunning() {
		// Define um timeout para não bloquear indefinidamente
		m.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := m.conn.ReadFromUDP(buffer)
		if err != nil {
			// Ignora erros de timeout
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("[%s] Erro ao ler dados: %v", m.config.MachineName, err)
			continue
		}

		// Processa os dados recebidos
		data := string(buffer[:n])
		log.Printf("[%s] Recebido de %s: %s", m.config.MachineName, addr, data)

		m.handleReceivedData(data)
	}
}

// Stop encerra a operação da máquina
// Fecha conexões e para timers
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

// handleReceivedData processa os dados recebidos pela rede
// Identifica se é um token ou pacote de dados e encaminha para o handler apropriado
func (m *Machine) handleReceivedData(data string) {
	m.updateLastActivity()

	// Verifica se é um pacote de token
	if message.IsTokenPacket(data) {
		m.handleToken()
		return
	}

	// Se não for token, tenta parsear como pacote de dados
	dataMsg, err := message.ParseDataPacket(data)
	if err != nil {
		log.Printf("[%s] Erro ao parsear pacote de dados: %v", m.config.MachineName, err)
		return
	}

	m.handleDataPacket(dataMsg)
}

// handleToken processa o recebimento de um token
// Atualiza o estado da máquina e agenda o processamento do token
func (m *Machine) handleToken() {
	log.Printf("[%s] Token recebido", m.config.MachineName)

	m.mutex.Lock()
	m.hasToken = true
	m.status.HasToken = true
	m.status.TokensProcessed++
	m.mutex.Unlock()

	// Cancela qualquer timer de token anterior
	if m.tokenTimeout != nil {
		m.tokenTimeout.Stop()
	}

	// Agenda o processamento do token após o tempo configurado
	m.tokenTimeout = time.AfterFunc(time.Duration(m.config.TokenTime)*time.Second, func() {
		m.processToken()
	})
}

// processToken é chamado quando o tempo de posse do token expira
// Verifica se há mensagens na fila para enviar ou passa o token adiante
func (m *Machine) processToken() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Verifica se ainda possui o token
	if !m.hasToken {
		return
	}

	// Se há mensagens na fila, envia a primeira
	if !m.queue.IsEmpty() {
		queuedMsg := m.queue.Peek()
		if queuedMsg != nil {
			// Cria um pacote de dados com a mensagem da fila
			dataMsg := message.CreateDataPacket(m.config.MachineName, queuedMsg.Destination, queuedMsg.Content)

			// Tratamento especial para mensagens broadcast
			if queuedMsg.Destination == "TODOS" {
				log.Printf("[%s] Enviando mensagem BROADCAST: %s", m.config.MachineName, queuedMsg.Content)
			} else {
				// Introduz erro com probabilidade configurada (exceto para broadcast)
				if dataMsg.IntroduceError(m.errorProbability) {
					log.Printf("[%s] Erro introduzido na mensagem para %s", m.config.MachineName, queuedMsg.Destination)
				}
			}

			// Marca que está aguardando resposta para esta mensagem
			m.waitingForData = true
			m.currentDataMsg = dataMsg

			// Envia o pacote e atualiza estatísticas
			m.sendPacket(dataMsg.RawData)
			m.status.MessagesSent++

			log.Printf("[%s] Mensagem enviada para %s: %s", m.config.MachineName, queuedMsg.Destination, queuedMsg.Content)
		}
	} else {
		// Se não há mensagens, passa o token adiante
		log.Printf("[%s] Fila vazia, passando token", m.config.MachineName)
		m.passToken()
	}
}

// handleDataPacket processa um pacote de dados recebido
// Determina se a mensagem é para esta máquina, se é uma mensagem retornada,
// ou se deve ser encaminhada
func (m *Machine) handleDataPacket(dataMsg *message.DataMessage) {
	log.Printf("[%s] Pacote de dados recebido: %s", m.config.MachineName, dataMsg.String())

	// Verifica se a mensagem é para esta máquina ou é broadcast
	if dataMsg.Destination == m.config.MachineName || dataMsg.Destination == "TODOS" {
		m.handleMessageForThisMachine(dataMsg)
	} else if dataMsg.Origin == m.config.MachineName {
		m.handleReturnedMessage(dataMsg)
	} else {
		m.forwardMessage(dataMsg)
	}
}

// handleMessageForThisMachine processa uma mensagem destinada a esta máquina
// Verifica a integridade usando CRC e envia ACK/NAK apropriado
func (m *Machine) handleMessageForThisMachine(dataMsg *message.DataMessage) {
	m.status.MessagesReceived++

	// Tratamento especial para mensagens broadcast
	if dataMsg.Destination == "TODOS" {
		log.Printf("[%s] Mensagem BROADCAST recebida de %s: %s", m.config.MachineName, dataMsg.Origin, dataMsg.Message)

		// Se a origem do broadcast é esta própria máquina, significa que completou o ciclo
		if dataMsg.Origin == m.config.MachineName {
			m.mutex.Lock()
			m.waitingForData = false
			m.currentDataMsg = nil
			m.queue.RemoveFirstMessage()
			m.passToken()
			m.mutex.Unlock()
			return
		}

		// Encaminha o broadcast para a próxima máquina
		m.forwardMessage(dataMsg)
		return
	}

	// Para mensagens unicast, verifica a integridade usando CRC
	if dataMsg.VerifyIntegrity() {
		log.Printf("[%s] Mensagem recebida de %s: %s", m.config.MachineName, dataMsg.Origin, dataMsg.Message)
		dataMsg.SetControl(message.ControlACK) // Envia ACK se íntegra
	} else {
		log.Printf("[%s] Erro detectado na mensagem de %s", m.config.MachineName, dataMsg.Origin)
		dataMsg.SetControl(message.ControlNAK) // Envia NAK se corrompida
		m.status.ErrorsDetected++
	}

	// Envia a resposta (ACK/NAK) de volta para a origem
	m.sendPacket(dataMsg.RawData)
}

// handleReturnedMessage processa uma mensagem que retornou à sua origem
// Analisa o campo de controle (ACK/NAK) e toma a ação apropriada
func (m *Machine) handleReturnedMessage(dataMsg *message.DataMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Caso especial para broadcast que completou o ciclo
	if dataMsg.Destination == "TODOS" {
		log.Printf("[%s] Mensagem BROADCAST completou o ciclo", m.config.MachineName)
		m.queue.RemoveFirstMessage()
		m.waitingForData = false
		m.currentDataMsg = nil
		m.passToken()
		return
	}

	// Verifica se estava esperando resposta para alguma mensagem
	if !m.waitingForData || m.currentDataMsg == nil {
		log.Printf("[%s] Mensagem retornada inesperada", m.config.MachineName)
		return
	}

	// Limpa o estado de espera
	m.waitingForData = false
	m.currentDataMsg = nil

	// Processa o campo de controle da mensagem
	switch dataMsg.Control {
	case message.ControlACK:
		// Mensagem recebida com sucesso, remove da fila
		log.Printf("[%s] ACK recebido para mensagem para %s", m.config.MachineName, dataMsg.Destination)
		m.queue.RemoveFirstMessage()

	case message.ControlNAK:
		// Erro detectado, incrementa contador de tentativas para retransmissão
		log.Printf("[%s] NAK recebido para mensagem para %s - será retransmitida", m.config.MachineName, dataMsg.Destination)
		m.queue.IncrementRetries()

	case message.ControlMachineNotExists:
		// Destinatário não existe, remove da fila
		log.Printf("[%s] Máquina %s não existe ou está desligada", m.config.MachineName, dataMsg.Destination)
		m.queue.RemoveFirstMessage()
	}

	// Passa o token adiante após processar a resposta
	m.passToken()
}

// forwardMessage encaminha uma mensagem para a próxima máquina na rede
// Usado quando a mensagem não é para esta máquina
func (m *Machine) forwardMessage(dataMsg *message.DataMessage) {
	log.Printf("[%s] Repassando mensagem de %s para %s", m.config.MachineName, dataMsg.Origin, dataMsg.Destination)
	m.sendPacket(dataMsg.RawData)
}

// passToken libera o token e o envia para a próxima máquina na rede
func (m *Machine) passToken() {
	// Atualiza o estado para indicar que não possui mais o token
	m.hasToken = false
	m.status.HasToken = false

	// Cria e envia o pacote de token
	tokenPacket := message.CreateTokenPacket()
	m.sendPacket(tokenPacket)

	log.Printf("[%s] Token enviado para próxima máquina", m.config.MachineName)
}

// sendPacket envia um pacote para a próxima máquina na rede
// Utiliza o endereço configurado em NextMachineAddr
func (m *Machine) sendPacket(data string) error {
	// Resolve o endereço UDP da próxima máquina
	addr, err := net.ResolveUDPAddr("udp", m.config.NextMachineAddr)
	if err != nil {
		return fmt.Errorf("erro ao resolver endereço: %v", err)
	}

	// Envia os dados via UDP
	_, err = m.conn.WriteToUDP([]byte(data), addr)
	if err != nil {
		return fmt.Errorf("erro ao enviar pacote: %v", err)
	}

	return nil
}

// generateInitialToken gera e envia o token inicial para a rede
// Chamado apenas pela máquina configurada para gerar o token
func (m *Machine) generateInitialToken() {
	log.Printf("[%s] Gerando token inicial", m.config.MachineName)

	// Atualiza estatísticas
	m.mutex.Lock()
	m.status.TokensGenerated++
	m.mutex.Unlock()

	// Cria e envia o pacote de token
	tokenPacket := message.CreateTokenPacket()
	err := m.sendPacket(tokenPacket)
	if err != nil {
		log.Printf("[%s] Erro ao enviar token inicial: %v", m.config.MachineName, err)
	}
}

// QueueMessage adiciona uma mensagem à fila para envio posterior
// Chamado pela interface de usuário quando uma mensagem deve ser enviada
func (m *Machine) QueueMessage(destination, content string) error {
	return m.queue.Enqueue(destination, content)
}

// GetStatus retorna o status atual da máquina
// Usado para exibir informações na interface de usuário
func (m *Machine) GetStatus() MachineStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Cria uma cópia do status para evitar condições de corrida
	status := *m.status
	status.QueueSize = m.queue.Size()
	status.LastActivity = m.lastActivity

	return status
}

// GetMessageQueue retorna todas as mensagens na fila
// Usado para exibir a fila na interface de usuário
func (m *Machine) GetMessageQueue() []*message.QueuedMessage {
	return m.queue.GetAll()
}

// GenerateToken força a geração de um novo token
// Só pode ser chamado se a máquina não possuir o token atualmente
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
// Usado no loop principal para determinar se deve continuar
func (m *Machine) isRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// updateLastActivity atualiza o timestamp da última atividade
// Chamado sempre que há comunicação na rede
func (m *Machine) updateLastActivity() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lastActivity = time.Now()
}

// tokenWatchdog monitora a circulação do token na rede
// Se o token não for visto por um tempo máximo, gera um novo token
// Implementa um mecanismo de recuperação de falhas na rede
func (m *Machine) tokenWatchdog() {
	// Apenas a máquina autorizada a gerar tokens executa o watchdog
	if !m.config.GeneratesToken {
		return
	}

	// Calcula o tempo máximo esperado para circulação do token
	// Considera o tempo do token multiplicado pelo número estimado de máquinas
	// e adiciona uma margem de segurança
	maxTokenCirculationTime := time.Duration(m.config.TokenTime*3*2+3) * time.Second
	ticker := time.NewTicker(maxTokenCirculationTime)
	defer ticker.Stop()

	// Inicializa o timestamp da última vez que o token foi visto
	lastTokenSeen := time.Now()

	for {
		select {
		// A cada tick, verifica se o token está circulando
		case <-ticker.C:
			m.mutex.RLock()
			hasToken := m.hasToken
			running := m.running
			timeSinceLastToken := time.Since(lastTokenSeen)
			m.mutex.RUnlock()

			// Se a máquina não está mais em execução, encerra o watchdog
			if !running {
				return
			}

			// Se o token não foi visto por muito tempo e esta máquina não o possui,
			// assume que o token foi perdido e gera um novo
			if timeSinceLastToken > maxTokenCirculationTime && !hasToken {
				log.Printf("[%s] Token perdido! (último visto há %v) Gerando novo token...",
					m.config.MachineName, timeSinceLastToken)
				m.generateInitialToken()
				lastTokenSeen = time.Now()
			}

		// No tempo entre ticks, verifica continuamente se a máquina possui o token
		default:
			m.mutex.RLock()
			if m.hasToken {
				lastTokenSeen = time.Now()
			}
			m.mutex.RUnlock()

			// Pequena pausa para não consumir CPU desnecessariamente
			time.Sleep(1 * time.Second)
		}
	}
}
