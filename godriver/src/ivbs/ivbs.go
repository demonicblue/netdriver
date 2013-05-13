package ivbs

import "encoding/binary"

const (
	OP_GREETING		uint32 = 1
	OP_KEEPALIVE	uint32 = 2

	OP_LOGIN		uint32 = 100

	OP_READ			uint32 = 300
	OP_WRITE		uint32 = 301

	OP_LIST_PROXIES	uint32 = 211
)

const (
    STATUS_OK                = 0
    STATUS_NOT_FOUND         = 1
    STATUS_ALREADY_EXISTS    = 2
    STATUS_NOT_LOGGEDIN      = 3
    STATUS_PERMISSION_DENIED = 4
    STATUS_READ_ONLY         = 5
    STATUS_NOT_READY         = 6

    STATUS_TEMPORARY_ERROR   = 10
    STATUS_PERMANENT_ERROR   = 11
    STATUS_USE_ANOTHER_PROXY = 12
    STATUS_INVALID_REQUEST   = 13
    STATUS_EXPIRED           = 14
)

type IvbsLogin struct {
    Name         string
    PasswordHash string
}
const LEN_USERNAME      = 32
const LEN_PASSWORD_HASH = 128

type IvbsPacket struct {
	SessionId [32]byte
	Op uint32
	Status int8
	DataLength uint32
	Sequence uint32
}

func IvbsStructToSlice(packet *IvbsPacket) ([]byte) {
	data := make([]byte, 45)
	
	copy(data[:32], packet.SessionId[:])
	binary.BigEndian.PutUint32(data[32:36], packet.Op)
	data[36] = byte(packet.Status)
	binary.BigEndian.PutUint32(data[37:41], packet.DataLength)
	binary.BigEndian.PutUint32(data[41:], packet.Sequence)
	
	return data
}

func LoginStructToSlice(packet *IvbsLogin) ([]byte) {
	data := make([]byte, LEN_USERNAME + LEN_PASSWORD_HASH)
	
	copy(data[:LEN_USERNAME], []byte(packet.Name))
	copy(data[LEN_USERNAME:], []byte(packet.PasswordHash))
	
	return data
}