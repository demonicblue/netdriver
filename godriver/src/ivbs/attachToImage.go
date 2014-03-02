package ivbs

import (
	"encoding/binary"
)

type AttachToImage struct {
	Name     string
	Size     uint64
	ReadOnly bool
}

func (attach *AttachToImage) Write(b []byte) (n int) {
	n += copy(b[:LEN_IMAGENAME], []byte(attach.Name))
	return n
}

func NewAttach(sequence SequenceGetter, image string) (packet *Packet) {
	packet = NewPacket()

	sequence.WriteSession(packet.SessionId[:])

	packet.Op = OP_ATTACH_TO_IMAGE
	packet.DataLen = LEN_IMAGENAME
	packet.Sequence = sequence.GetSequence()

	tmp := new(AttachToImage)
	tmp.Name = image

	packet.DataPacket = tmp

	return packet
}

func AttachFromSlice(packet *Packet) (attach *AttachToImage) {
	attach = new(AttachToImage)
	b := packet.DataSlice[LEN_HEADER_PACKET:]

	attach.Name = string(b[:LEN_IMAGENAME])
	attach.Size = binary.BigEndian.Uint64(b[LEN_IMAGENAME : LEN_IMAGENAME+8])

	return attach
}
