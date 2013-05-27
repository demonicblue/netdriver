package ivbs

import (
	"encoding/binary"
	"fmt"
)

const (
	OP_GREETING		uint32 = 1
	OP_KEEPALIVE	uint32 = 2

	OP_LOGIN		uint32 = 100

	OP_READ			uint32 = 300
	OP_WRITE		uint32 = 301

	OP_ATTACH_TO_IMAGE   uint32 = 200

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

const LEN_HEADER_PACKET = 48

const LEN_USERNAME      = 32
const LEN_PASSWORD_HASH = 128

const (
	LEN_IMAGENAME     = 256
)

const (
	LEN_SESSIONID     = 32
	LEN_OP            = 4
	LEN_STATUS        = 4
	LEN_DATALEN       = 4
	LEN_SEQ           = 4
	LEN_PACKET_HEADER = LEN_SESSIONID + LEN_OP + LEN_STATUS + LEN_DATALEN + LEN_SEQ
)

type Packet struct {
	SessionId []byte
	Op uint32
	Status uint32
	DataLen uint32
	Sequence uint32
	DataPacket PacketData
	DataSlice []byte
}

type PacketData interface {
	Write(b []byte) int
}

type SequenceGetter interface {
	GetSequence() uint32
	WriteSession(b []byte)
}

func NewPacket() (packet *Packet) {
	packet = new(Packet)
	packet.SessionId = make([]byte, LEN_SESSIONID)
	return packet
}

func (packet *Packet) Byteslice() (data []byte) {
	data = make([]byte, LEN_HEADER_PACKET + packet.DataLen)

	copy(data[:32], packet.SessionId[:])
	binary.BigEndian.PutUint32(data[32:36], packet.Op)
	binary.BigEndian.PutUint32(data[36:40], packet.Status)
	binary.BigEndian.PutUint32(data[40:44], packet.DataLen)
	binary.BigEndian.PutUint32(data[44:48], packet.Sequence)

	if packet.DataLen > 0 {
		fmt.Printf("Trying to write to a bytslice of length %d\n", len(data))
		packet.DataPacket.Write( data[48:] )
	}

	packet.DataSlice = data

	return data
}

func IvbsSliceToStruct(data []byte) (*Packet) {
	packet := new(Packet)
	packet.SessionId = make([]byte, LEN_SESSIONID)
	
	copy(packet.SessionId[:], data[:32])
	packet.Op = binary.BigEndian.Uint32(data[32:36])
	packet.Status = binary.BigEndian.Uint32(data[36:40])
	packet.DataLen = binary.BigEndian.Uint32(data[40:44])
	packet.Sequence = binary.BigEndian.Uint32(data[44:48])
	
	return packet
}

func (packet *Packet) Debug() {
	fmt.Printf("%+v \n", packet)
}