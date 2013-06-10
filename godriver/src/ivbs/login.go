package ivbs

import (
	//"encoding/binary"
	"crypto/sha512"
	"fmt"
)

type Login struct {
	Name         string
	Password     string
	PasswordHash string
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

func LoginSliceToStruct(data []byte) *Login {
	packet := new(Login)

	packet.Name = string(data[:LEN_USERNAME])
	packet.PasswordHash = string(data[LEN_USERNAME : LEN_USERNAME+LEN_PASSWORD_HASH])

	return packet
}
