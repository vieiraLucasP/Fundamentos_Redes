package message

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"ring-network/pkg/crc"
)

const (
	TokenPacket = "1000"
	DataPacket  = "2000"
)

const (
	ControlMachineNotExists = "maquinanaoexiste"
	ControlACK              = "ACK"
	ControlNAK              = "NAK"
)

type QueuedMessage struct {
	Destination string
	Content     string
	Timestamp   time.Time
	Retries     int
}

type DataMessage struct {
	Type        string
	Origin      string
	Destination string
	Control     string
	CRC         string
	Message     string
	RawData     string
}

func NewQueuedMessage(destination, content string) *QueuedMessage {
	return &QueuedMessage{
		Destination: destination,
		Content:     content,
		Timestamp:   time.Now(),
		Retries:     0,
	}
}

func CreateDataPacket(origin, destination, message string) *DataMessage {
	dataForCRC := crc.CreateDataForCRC(origin, destination, message)
	crcValue := crc.CalculateCRC32String(dataForCRC)

	control := ControlMachineNotExists

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

func ParseDataPacket(data string) (*DataMessage, error) {
	if !strings.HasPrefix(data, DataPacket+";") {
		return nil, fmt.Errorf("não é um pacote de dados válido")
	}

	content := strings.TrimPrefix(data, DataPacket+";")

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

func IsTokenPacket(data string) bool {
	return strings.TrimSpace(data) == TokenPacket
}

func CreateTokenPacket() string {
	return TokenPacket
}

func (dm *DataMessage) VerifyIntegrity() bool {
	dataForCRC := crc.CreateDataForCRC(dm.Origin, dm.Destination, dm.Message)
	return crc.VerifyCRC32String(dataForCRC, dm.CRC)
}

func (dm *DataMessage) SetControl(control string) {
	dm.Control = control
	dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
		DataPacket, dm.Origin, dm.Destination, control, dm.CRC, dm.Message)
}

func (dm *DataMessage) IntroduceError(probability float64) bool {
	if rand.Float64() < probability {
		originalCRC := dm.CRC

		corruptedCRC := strconv.FormatUint(uint64(rand.Uint32()), 10)

		for corruptedCRC == originalCRC {
			corruptedCRC = strconv.FormatUint(uint64(rand.Uint32()), 10)
		}

		dm.CRC = corruptedCRC

		dm.RawData = fmt.Sprintf("%s;%s:%s:%s:%s:%s",
			DataPacket, dm.Origin, dm.Destination, dm.Control, dm.CRC, dm.Message)

		return true
	}
	return false
}

func (dm *DataMessage) String() string {
	return fmt.Sprintf("DataMessage{Origin: %s, Destination: %s, Control: %s, Message: %s}",
		dm.Origin, dm.Destination, dm.Control, dm.Message)
}

func (qm *QueuedMessage) String() string {
	return fmt.Sprintf("QueuedMessage{Destination: %s, Content: %s, Retries: %d}",
		qm.Destination, qm.Content, qm.Retries)
}
