package utils

func GenerateId(clientType, clientId uint16) uint16 {
	return (clientType<<8 + clientId)
}
