package message

import (
	"strings"
	"testing"
)

func TestNewQueuedMessage(t *testing.T) {
	destination := "Bob"
	content := "Hello World"

	msg := NewQueuedMessage(destination, content)

	if msg.Destination != destination {
		t.Errorf("Destino incorreto: %s != %s", msg.Destination, destination)
	}

	if msg.Content != content {
		t.Errorf("Conteúdo incorreto: %s != %s", msg.Content, content)
	}

	if msg.Retries != 0 {
		t.Errorf("Tentativas deveria ser 0, obtido %d", msg.Retries)
	}
}

func TestCreateDataPacket(t *testing.T) {
	origin := "Alice"
	destination := "Bob"
	message := "Hello World"

	packet := CreateDataPacket(origin, destination, message)

	if packet.Type != DataPacket {
		t.Errorf("Tipo incorreto: %s != %s", packet.Type, DataPacket)
	}

	if packet.Origin != origin {
		t.Errorf("Origem incorreta: %s != %s", packet.Origin, origin)
	}

	if packet.Destination != destination {
		t.Errorf("Destino incorreto: %s != %s", packet.Destination, destination)
	}

	if packet.Control != ControlMachineNotExists {
		t.Errorf("Controle incorreto: %s != %s", packet.Control, ControlMachineNotExists)
	}

	if packet.Message != message {
		t.Errorf("Mensagem incorreta: %s != %s", packet.Message, message)
	}

	// Verificar formato do pacote
	if !strings.HasPrefix(packet.RawData, "2000;") {
		t.Error("Pacote não inicia com '2000;'")
	}
}

func TestParseDataPacket(t *testing.T) {
	// Criar um pacote válido
	rawData := "2000;Alice:Bob:maquinanaoexiste:12345:Hello World"

	packet, err := ParseDataPacket(rawData)
	if err != nil {
		t.Fatalf("Erro ao parsear pacote válido: %v", err)
	}

	if packet.Origin != "Alice" {
		t.Errorf("Origem incorreta: %s", packet.Origin)
	}

	if packet.Destination != "Bob" {
		t.Errorf("Destino incorreto: %s", packet.Destination)
	}

	if packet.Control != "maquinanaoexiste" {
		t.Errorf("Controle incorreto: %s", packet.Control)
	}

	if packet.CRC != "12345" {
		t.Errorf("CRC incorreto: %s", packet.CRC)
	}

	if packet.Message != "Hello World" {
		t.Errorf("Mensagem incorreta: %s", packet.Message)
	}
}

func TestParseDataPacketInvalid(t *testing.T) {
	// Testar pacote inválido
	invalidData := "1000"

	_, err := ParseDataPacket(invalidData)
	if err == nil {
		t.Error("Deveria retornar erro para pacote inválido")
	}

	// Testar formato incompleto
	incompleteData := "2000;Alice:Bob"

	_, err = ParseDataPacket(incompleteData)
	if err == nil {
		t.Error("Deveria retornar erro para pacote incompleto")
	}
}

func TestIsTokenPacket(t *testing.T) {
	// Testar token válido
	if !IsTokenPacket("1000") {
		t.Error("Deveria reconhecer '1000' como token")
	}

	if !IsTokenPacket(" 1000 ") {
		t.Error("Deveria reconhecer '1000' com espaços como token")
	}

	// Testar não-token
	if IsTokenPacket("2000") {
		t.Error("Não deveria reconhecer '2000' como token")
	}

	if IsTokenPacket("2000;Alice:Bob:ACK:12345:Hello") {
		t.Error("Não deveria reconhecer pacote de dados como token")
	}
}

func TestCreateTokenPacket(t *testing.T) {
	token := CreateTokenPacket()

	if token != TokenPacket {
		t.Errorf("Token incorreto: %s != %s", token, TokenPacket)
	}
}

func TestVerifyIntegrity(t *testing.T) {
	// Criar pacote com CRC correto
	packet := CreateDataPacket("Alice", "Bob", "Hello World")

	// Verificar integridade
	if !packet.VerifyIntegrity() {
		t.Error("Verificação de integridade falhou para pacote válido")
	}

	// Corromper CRC
	packet.CRC = "99999"

	// Verificar que agora falha
	if packet.VerifyIntegrity() {
		t.Error("Verificação de integridade passou para pacote corrompido")
	}
}

func TestSetControl(t *testing.T) {
	packet := CreateDataPacket("Alice", "Bob", "Hello World")

	// Mudar controle para ACK
	packet.SetControl(ControlACK)

	if packet.Control != ControlACK {
		t.Errorf("Controle não foi alterado: %s != %s", packet.Control, ControlACK)
	}

	// Verificar se RawData foi atualizado
	if !strings.Contains(packet.RawData, ":ACK:") {
		t.Error("RawData não foi atualizado com novo controle")
	}
}
