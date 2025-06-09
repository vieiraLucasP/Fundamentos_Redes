package crc

import (
	"hash/crc32"
	"strconv"
)

func CalculateCRC32(data string) uint32 {
	return crc32.ChecksumIEEE([]byte(data))
}

func CalculateCRC32String(data string) string {
	crc := CalculateCRC32(data)
	return strconv.FormatUint(uint64(crc), 10)
}

func VerifyCRC32(data string, expectedCRC uint32) bool {
	calculated := CalculateCRC32(data)
	return calculated == expectedCRC
}

func VerifyCRC32String(data string, expectedCRCStr string) bool {
	expectedCRC, err := strconv.ParseUint(expectedCRCStr, 10, 32)
	if err != nil {
		return false
	}
	return VerifyCRC32(data, uint32(expectedCRC))
}

func CreateDataForCRC(origin, destination, message string) string {
	return origin + ":" + destination + ":" + message
}
