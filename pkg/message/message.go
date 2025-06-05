package message

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"ring-network/pkg/crc"
)

// Tipos de pacote
const (
	TokenPacket = "1000"
	DataPacket  = "2000"
)

// Estados de controle
const (
	ControlMachineNotExists = "maquinanaoexiste"
	ControlACK              = "ACK"
	ControlNAK              = "NAK"
)

// QueuedMessage representa uma mensagem na fila
type QueuedMessage struct {
	Destination string
	Content     string
	Timestamp   time.Time
	Retries     int
}

// DataMessage representa um pacote de dados
type DataMessage struct {
	Type        string
	Origin      string
	Destination string
	Control     string
	CRC         string
	Message     string
	RawData     string
}

// NewQueuedMessage cria uma nova mensagem para a fila
func NewQueuedMessage(destination, content string) *QueuedMessage {
	return &QueuedMessage{
		Destination: destination,
		Content:     content,
		Timestamp:   time.Now(),
		Retries:     0,
	}
}

// CreateDataPacket cria um pacote de dados
func CreateDataPacket(origin, destination, message string) *DataMessage {
	// Dados para cálculo do CRC (sem o controle)
	dataForCRC := crc.CreateDataForCRC(origin, destination, message)
	crcValue := crc.CalculateCRC32String(dataForCRC)

	control := ControlMachineNotExists

	// Formato: 2000;origem:destino:controle:CRC:mensagem
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

// ParseDataPacket faz o parse de um pacote de dados recebido
func ParseDataPacket(data string) (*DataMessage, error) {
	// Verificar se é um pacote de dados
	if !strings.HasPrefix(data, DataPacket+";") {
		return nil, fmt.Errorf("não é um pacote de dados válido")
	}

	// Remover o prefixo "2000;"
	content := strings.TrimPrefix(data, DataPacket+";")

	// Dividir por ':' - origem:destino:controle:CRC:mensagem
	parts := strings.SplitN(content, ":", 5)
	if len(parts) != 5 {
		return nil, fmt.Errorf("formato de pacote inválido: esperado 5 partes, obtido %d", len(parts))
	}

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

// IsTokenPacket verifica se os dados são um token
func IsTokenPacket(data string) bool {
	return strings.TrimSpace(data) == TokenPacket
}

// CreateTokenPacket cria um pacote de token
func CreateTokenPacket() string {
	return TokenPacket
}

// VerifyIntegrity verifica a integridade da mensagem usando CRC32
func (dm *DataMessage) VerifyIntegrity() bool {
	// Recalcular CRC com os dados originais (sem o controle atual)
	dataForCRC := crc.CreateDataForCRC(dm.Origin, dm.Destination, dm.Message)
	return crc.VerifyCRC32String(dataForCRC, dm.CRC)
}

// SetControl altera o estado de controle da mensagem
func (dm *DataMessage) SetControl(control string) {
	dm.Control = control
	// Recriar o raw data com o novo controle
	dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
		DataPacket, dm.Origin, dm.Destination, control, dm.CRC, dm.Message)
}

// IntroduceError introduz erro aleatório na mensagem (módulo de falhas)
func (dm *DataMessage) IntroduceError(probability float64) bool {
	if rand.Float64() < probability {
		// Corromper o CRC para simular erro de transmissão
		originalCRC := dm.CRC

		// Gerar um CRC inválido
		corruptedCRC := strconv.FormatUint(uint64(rand.Uint32()), 10)

		// Garantir que o CRC corrompido seja diferente do original
		for corruptedCRC == originalCRC {
			corruptedCRC = strconv.FormatUint(uint64(rand.Uint32()), 10)
		}

		dm.CRC = corruptedCRC

		// Recriar raw data com CRC corrompido
		dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
			DataPacket, dm.Origin, dm.Destination, dm.Control, dm.CRC, dm.Message)

		return true // Erro introduzido
	}
	return false // Nenhum erro introduzido
}

// String retorna uma representação em string da mensagem
func (dm *DataMessage) String() string {
	// Mostrar o conteúdo apenas para mensagens broadcast
	if dm.Destination == "TODOS" {
		return fmt.Sprintf("DataMessage{Origin: %s, Destination: %s, Control: %s, Message: %s}",
			dm.Origin, dm.Destination, dm.Control, dm.Message)
	}
	
	// Para mensagens privadas, não mostrar o conteúdo
	return fmt.Sprintf("DataMessage{Origin: %s, Destination: %s, Control: %s, Message: <conteúdo privado>}",
		dm.Origin, dm.Destination, dm.Control)
}

// String retorna uma representação em string da mensagem na fila
func (qm *QueuedMessage) String() string {
	return fmt.Sprintf("QueuedMessage{Destination: %s, Content: %s, Retries: %d}",
		qm.Destination, qm.Content, qm.Retries)
}
