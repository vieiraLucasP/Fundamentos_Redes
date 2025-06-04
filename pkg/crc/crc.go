package crc

import (
	"hash/crc32"
	"strconv"
)

// CalculateCRC32 calcula o CRC32 de uma string
func CalculateCRC32(data string) uint32 {
	return crc32.ChecksumIEEE([]byte(data))
}

// CalculateCRC32String calcula o CRC32 e retorna como string
func CalculateCRC32String(data string) string {
	crc := CalculateCRC32(data)
	return strconv.FormatUint(uint64(crc), 10)
}

// VerifyCRC32 verifica se o CRC32 de uma string corresponde ao esperado
func VerifyCRC32(data string, expectedCRC uint32) bool {
	calculated := CalculateCRC32(data)
	return calculated == expectedCRC
}

// VerifyCRC32String verifica usando strings
func VerifyCRC32String(data string, expectedCRCStr string) bool {
	expectedCRC, err := strconv.ParseUint(expectedCRCStr, 10, 32)
	if err != nil {
		return false
	}
	return VerifyCRC32(data, uint32(expectedCRC))
}

// CreateDataForCRC cria a string de dados para cálculo de CRC
// Usado para criar a string completa antes de calcular o CRC
func CreateDataForCRC(origin, destination, message string) string {
	return origin + ":" + destination + ":" + message
}
