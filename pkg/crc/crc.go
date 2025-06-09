package crc

import (
	"hash/crc32"
	"strconv"
)

// CalculateCRC32 calcula o valor CRC32 (IEEE) para uma string fornecida
// Utiliza a implementação padrão do pacote hash/crc32 da biblioteca Go
func CalculateCRC32(data string) uint32 {
	return crc32.ChecksumIEEE([]byte(data))
}

// CalculateCRC32String calcula o CRC32 e converte o resultado para string
// Útil para incluir o CRC em mensagens de texto
func CalculateCRC32String(data string) string {
	crc := CalculateCRC32(data)
	return strconv.FormatUint(uint64(crc), 10)
}

// VerifyCRC32 verifica se o CRC32 calculado para os dados corresponde ao valor esperado
// Utilizado para verificar a integridade dos dados recebidos
func VerifyCRC32(data string, expectedCRC uint32) bool {
	calculated := CalculateCRC32(data)
	return calculated == expectedCRC
}

// VerifyCRC32String verifica o CRC32 quando o valor esperado está em formato string
// Converte a string para uint32 antes de fazer a comparação
func VerifyCRC32String(data string, expectedCRCStr string) bool {
	expectedCRC, err := strconv.ParseUint(expectedCRCStr, 10, 32)
	if err != nil {
		return false
	}
	return VerifyCRC32(data, uint32(expectedCRC))
}

// CreateDataForCRC cria uma string padronizada para cálculo do CRC
// Concatena origem, destino e mensagem com separadores para garantir consistência
func CreateDataForCRC(origin, destination, message string) string {
	return origin + ":" + destination + ":" + message
}
