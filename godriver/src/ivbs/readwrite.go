package ivbs

import (
	"encoding/binary"
)

type Read struct {
	Offset  uint64 //<- client->ivbs
	DataLen uint64 //<- client->ivbs
	Data    []byte //<- ivbs->client
}
type Write struct {
	Offset  uint64 //<- client->ivbs
	Data    []byte //<- client->ivbs
}

func (read *Read) Write(b []byte) (n int) {
	binary.BigEndian.PutUint64(b[:8], read.Offset)
	binary.BigEndian.PutUint64(b[8:16], read.DataLen)

	return 16
}

func NewRead(sequence SequenceGetter, offset, dataLen uint64) (packet *Packet) {
	packet = NewPacket()

	sequence.WriteSession(packet.SessionId[:])

	packet.Op = OP_READ
	packet.Sequence = sequence.GetSequence()
	packet.DataLen = 16

	tmp := new(Read)
	tmp.Offset = offset
	tmp.DataLen = dataLen
	packet.DataPacket = tmp

	return packet
}

func (write *Write) Write(b []byte) (n int) {
	binary.BigEndian.PutUint64(b[:8], write.Offset)
	n += copy(b[8:], write.Data)

	return n+8
}

func NewWrite(sequence SequenceGetter, offset uint64, length uint32, b []byte) (packet *Packet) {
	packet = NewPacket()

	sequence.WriteSession(packet.SessionId[:])

	packet.Op = OP_WRITE
	packet.Sequence = sequence.GetSequence()
	packet.DataLen = 8 + length

	tmp := new(Write)
	tmp.Offset = offset
	tmp.Data = b

	return packet
}

