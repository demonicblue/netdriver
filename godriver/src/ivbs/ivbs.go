package ivbs

import "encoding/binary"
import "crypto/sha512"
import "fmt"

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

const LEN_HEADER_PACKET = 48

type Login struct {
    Name string
    Password string
    PasswordHash string
}
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

type AttachToImage struct {
	Name string
	Size uint64
	ReadOnly bool
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

func (attach *AttachToImage) Write(b []byte) (n int) {
	n += copy(b[:LEN_IMAGENAME], []byte(attach.Name))
	return n
}

func (login *Login) Write(b []byte) (n int) {
	n += copy(b[:LEN_USERNAME], []byte(login.Name))

	c := sha512.New()
	c.Write([]byte(login.Password))

	str := fmt.Sprintf("%x", c.Sum(nil))

	n += copy(b[LEN_USERNAME:LEN_USERNAME+LEN_PASSWORD_HASH], []byte(str))

	return n
}

/*
*	Create new login packet with all values set, ready for transmission.
*/
func NewLogin(sequence SequenceGetter, name, password string) (packet *Packet) {
	packet = NewPacket()

	sequence.WriteSession(packet.SessionId[:])

	packet.Op = OP_LOGIN
	packet.DataLen = LEN_USERNAME + LEN_PASSWORD_HASH
	packet.Sequence = sequence.GetSequence()

	tmp := new(Login)
	tmp.Name = name
	tmp.Password = password
	packet.DataPacket = tmp

	return packet
}

func (packet *Packet) Byteslice() (data []byte) {
	data = make([]byte, LEN_HEADER_PACKET + packet.DataLen)

	copy(data[:32], packet.SessionId[:])
	binary.BigEndian.PutUint32(data[32:36], packet.Op)
	binary.BigEndian.PutUint32(data[36:40], packet.Status)
	binary.BigEndian.PutUint32(data[40:44], packet.DataLen)
	binary.BigEndian.PutUint32(data[44:48], packet.Sequence)

	packet.DataPacket.Write( data[48:] )

	return data
}

func IvbsStructToSlice(packet *Packet) ([]byte) {
	data := make([]byte, LEN_HEADER_PACKET)
	
	copy(data[:32], packet.SessionId[:])
	binary.BigEndian.PutUint32(data[32:36], packet.Op)
	binary.BigEndian.PutUint32(data[36:40], packet.Status)
	binary.BigEndian.PutUint32(data[40:44], packet.DataLen)
	binary.BigEndian.PutUint32(data[44:48], packet.Sequence)
	
	return data
}

func LoginStructToSlice(packet *Login) ([]byte) {
	data := make([]byte, LEN_USERNAME + LEN_PASSWORD_HASH)
	
	c := sha512.New()
	c.Write([]byte(packet.PasswordHash))
	str := fmt.Sprintf("%x", c.Sum(nil))
	
	copy(data[:LEN_USERNAME], []byte(packet.Name))
	copy(data[LEN_USERNAME:], []byte(str))
	//copy(data[LEN_USERNAME:], packet.PasswordHash)
	
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

func LoginSliceToStruct(data []byte) (*Login) {
	packet := new(Login)
	
	packet.Name = string(data[:LEN_USERNAME])
	packet.PasswordHash = string(data[LEN_USERNAME:LEN_USERNAME+LEN_PASSWORD_HASH])
	
	return packet
}