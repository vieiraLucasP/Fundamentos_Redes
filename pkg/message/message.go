package message

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"ring-network/pkg/crc"
)

// Constantes para identificação dos tipos de pacotes
const (
	TokenPacket = "1000" // Identificador do pacote de token
	DataPacket  = "2000" // Identificador do pacote de dados
)

// Constantes para os campos de controle das mensagens
const (
	ControlMachineNotExists = "maquinanaoexiste" // Indica que a máquina de destino não existe
	ControlACK              = "ACK"              // Confirmação positiva de recebimento
	ControlNAK              = "NAK"              // Confirmação negativa (erro detectado)
)

// QueuedMessage representa uma mensagem na fila para envio
type QueuedMessage struct {
	Destination string    // Destino da mensagem
	Content     string    // Conteúdo da mensagem
	Timestamp   time.Time // Momento de criação da mensagem
	Retries     int       // Número de tentativas de envio
}

// DataMessage representa um pacote de dados para transmissão na rede
type DataMessage struct {
	Type        string // Tipo do pacote (2000 para dados)
	Origin      string // Origem da mensagem
	Destination string // Destino da mensagem
	Control     string // Campo de controle (ACK, NAK, etc.)
	CRC         string // Valor CRC32 para verificação de integridade
	Message     string // Conteúdo da mensagem
	RawData     string // Representação em string do pacote completo
}

// NewQueuedMessage cria uma nova mensagem para a fila de envio
func NewQueuedMessage(destination, content string) *QueuedMessage {
	return &QueuedMessage{
		Destination: destination,
		Content:     content,
		Timestamp:   time.Now(),
		Retries:     0,
	}
}

// CreateDataPacket cria um novo pacote de dados para envio na rede
// Calcula o CRC32 para verificação de integridade
func CreateDataPacket(origin, destination, message string) *DataMessage {
	dataForCRC := crc.CreateDataForCRC(origin, destination, message)
	crcValue := crc.CalculateCRC32String(dataForCRC)

	// Inicialmente, o campo de controle indica que a máquina não existe
	// Este valor será alterado pelo destinatário ao receber a mensagem
	control := ControlMachineNotExists

	// Formata o pacote completo como string
	rawData := fmt.Sprintf("%s;%s:%s:%s:%s:%s",
		DataPacket, origin, destination, control, crcValue, message)

	return &DataMessage{
		Type:        DataPacket,
		Origin:      origin,
		Destination: destination,
		Control:     control,
		CRC:         crcValue,
		Message:     message,
		RawData:     rawData,
	}
}

// ParseDataPacket analisa uma string recebida e converte para um objeto DataMessage
// Retorna erro se o formato não for válido
func ParseDataPacket(data string) (*DataMessage, error) {
	// Verifica se começa com o identificador de pacote de dados
	if !strings.HasPrefix(data, DataPacket+";") {
		return nil, fmt.Errorf("não é um pacote de dados válido")
	}

	// Remove o prefixo do tipo de pacote
	content := strings.TrimPrefix(data, DataPacket+";")

	// Divide o conteúdo em partes usando ":" como separador
	parts := strings.SplitN(content, ":", 5)
	if len(parts) != 5 {
		return nil, fmt.Errorf("formato de pacote inválido: esperado 5 partes, obtido %d", len(parts))
	}

	// Cria e retorna o objeto DataMessage
	return &DataMessage{
		Type:        DataPacket,
		Origin:      parts[0],
		Destination: parts[1],
		Control:     parts[2],
		CRC:         parts[3],
		Message:     parts[4],
		RawData:     data,
	}, nil
}

// IsTokenPacket verifica se uma string recebida é um pacote de token
func IsTokenPacket(data string) bool {
	return strings.TrimSpace(data) == TokenPacket
}

// CreateTokenPacket cria um novo pacote de token
func CreateTokenPacket() string {
	return TokenPacket
}

// VerifyIntegrity verifica a integridade da mensagem usando CRC32
// Retorna true se o CRC calculado corresponder ao CRC armazenado na mensagem
func (dm *DataMessage) VerifyIntegrity() bool {
	dataForCRC := crc.CreateDataForCRC(dm.Origin, dm.Destination, dm.Message)
	return crc.VerifyCRC32String(dataForCRC, dm.CRC)
}

// SetControl atualiza o campo de controle da mensagem e recria o pacote raw
func (dm *DataMessage) SetControl(control string) {
	dm.Control = control
	dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
		DataPacket, dm.Origin, dm.Destination, control, dm.CRC, dm.Message)
}

// IntroduceError introduz um erro na mensagem com uma probabilidade definida
// Modifica o CRC para simular corrupção de dados
func (dm *DataMessage) IntroduceError(probability float64) bool {
	if rand.Float64() < probability {
		// Guarda o CRC original
		originalCRC := dm.CRC

		// Gera um novo CRC aleatório
		corruptedCRC := strconv.FormatUint(uint64(rand.Uint32()), 10)

		// Garante que o CRC corrompido seja diferente do original
		for corruptedCRC == originalCRC {
			corruptedCRC = strconv.FormatUint(uint64(rand.Uint32()), 10)
		}

		// Substitui o CRC pelo valor corrompido
		dm.CRC = corruptedCRC

		// Atualiza o pacote raw com o CRC corrompido
		dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
			DataPacket, dm.Origin, dm.Destination, dm.Control, dm.CRC, dm.Message)

		return true
	}
	return false
}

// String retorna uma representação em string do objeto DataMessage
func (dm *DataMessage) String() string {
	return fmt.Sprintf("DataMessage{Origin: %s, Destination: %s, Control: %s, Message: %s}",
		dm.Origin, dm.Destination, dm.Control, dm.Message)
}

// String retorna uma representação em string do objeto QueuedMessage
func (qm *QueuedMessage) String() string {
	return fmt.Sprintf("QueuedMessage{Destination: %s, Content: %s, Retries: %d}",
		qm.Destination, qm.Content, qm.Retries)
}
