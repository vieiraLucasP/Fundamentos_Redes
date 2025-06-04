package crc

import (
	"testing"
)

func TestCalculateCRC32(t *testing.T) {
	data := "Alice:Bob:Hello World"
	crc := CalculateCRC32(data)

	// Verificar se o CRC não é zero (seria muito improvável)
	if crc == 0 {
		t.Error("CRC32 calculado não deveria ser zero")
	}

	// Verificar se o mesmo dado gera o mesmo CRC
	crc2 := CalculateCRC32(data)
	if crc != crc2 {
		t.Errorf("CRC32 inconsistente: %d != %d", crc, crc2)
	}
}

func TestCalculateCRC32String(t *testing.T) {
	data := "Alice:Bob:Hello World"
	crcStr := CalculateCRC32String(data)

	// Verificar se retorna uma string não vazia
	if crcStr == "" {
		t.Error("CRC32 string não deveria estar vazia")
	}
}

func TestVerifyCRC32(t *testing.T) {
	data := "Alice:Bob:Hello World"
	crc := CalculateCRC32(data)

	// Verificar CRC correto
	if !VerifyCRC32(data, crc) {
		t.Error("Verificação de CRC32 falhou para dados corretos")
	}

	// Verificar CRC incorreto
	if VerifyCRC32(data, crc+1) {
		t.Error("Verificação de CRC32 passou para CRC incorreto")
	}
}

func TestVerifyCRC32String(t *testing.T) {
	data := "Alice:Bob:Hello World"
	crcStr := CalculateCRC32String(data)

	// Verificar CRC correto
	if !VerifyCRC32String(data, crcStr) {
		t.Error("Verificação de CRC32 string falhou para dados corretos")
	}

	// Verificar CRC incorreto
	if VerifyCRC32String(data, "12345") {
		t.Error("Verificação de CRC32 string passou para CRC incorreto")
	}

	// Verificar CRC string inválida
	if VerifyCRC32String(data, "invalid") {
		t.Error("Verificação de CRC32 string passou para string inválida")
	}
}

func TestCreateDataForCRC(t *testing.T) {
	origin := "Alice"
	destination := "Bob"
	message := "Hello World"

	data := CreateDataForCRC(origin, destination, message)
	expected := "Alice:Bob:Hello World"

	if data != expected {
		t.Errorf("CreateDataForCRC retornou '%s', esperado '%s'", data, expected)
	}
}
