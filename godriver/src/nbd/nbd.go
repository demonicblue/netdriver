package nbd

import (
	"ioctl"
	"os"
	"encoding/binary"
)

// Values imported from nbd.h
const (
	NBD_SET_SOCK	= iota
	NBD_SET_BLKSIZE
	NBD_SET_SIZE
	NBD_DO_IT
	NBD_CLEAR_SOCK
	NBD_CLEAR_QUE
	NBD_PRINT_DEBUG
	NBD_SET_SIZE_BLOCKS
	NBD_DISCONNECT
	NBD_SET_TIMEOUT
	NBD_SET_FLAGS
)

const (
	NBD_CMD_READ	= iota
	NBD_CMD_WRITE
	NBD_CMD_DISC
	NBD_CMD_FLUSH
	NBD_CMD_TRIM
)

const (
	NBD_REQUEST_MAGIC	= 0x25609513
	NBD_REPLY_MAGIC		= 0x67446698
)

const (
	LEN_REQUEST_HEADER	= 28
	LEN_REPLY_HEADER	= 16
)

type Request struct {
	Magic uint32
	Cmd uint32
	Handle [8]byte
	From uint64
	Len uint32
	Data []byte
}

type Reply struct {
	Magic uint32
	Error uint32
	Handle [8]byte
	Data []byte
}

func NewReply(request *Request, b []byte) (reply *Reply) {
	// TODO Set magic
	reply  = new(Reply)

	reply.Magic = NBD_REPLY_MAGIC
	copy(reply.Handle[:], request.Handle[:])
	reply.Data = request.Data

	return reply
}

func NewRequest(b []byte) (request *Request) {
	request = new(Request)

	request.Magic = binary.BigEndian.Uint32(b[:4])
	request.Cmd = binary.BigEndian.Uint32(b[4:8])
	copy(request.Handle[:], b[8:16])
	request.From = binary.BigEndian.Uint64(b[16:24])
	request.Len = binary.BigEndian.Uint32(b[24:28])

	if request.Len > 0 {
		request.Data = make([]byte, request.Len)
	}

	return request
}

func (request *Request) Parse() {
	//ss
}

func (reply *Reply) Byteslice() (b []byte) {
	b = make([]byte, LEN_REPLY_HEADER + len(reply.Data))

	binary.BigEndian.PutUint32(b[:4], reply.Magic)
	binary.BigEndian.PutUint32(b[4:8], reply.Error)
	copy(b[8:], reply.Handle[:])

	copy(b[16:], reply.Data)

	return b
}

// Send command to ioctl
func Call(fd, req, data int) error {
	errno := ioctl.Call(uintptr(fd), int(ioctl.IO(0xab, int32(req))), uintptr(data))
	if errno != 0 {
		err := os.NewSyscallError("SYS_IOCTL", errno)
		return err
	}
	return nil
}

func Call2(fd uintptr, req, data int) error {
	errno := ioctl.Call(fd, int(ioctl.IO(0xab, int32(req))), uintptr(data))
	if errno != 0 {
		err := os.NewSyscallError("SYS_IOCTL", errno)
		return err
	}
	return nil
}

func CallUint64(fd uintptr, req int, data uint64) error {
	errno := ioctl.Call(fd, int(ioctl.IO(0xab, int32(req))), uintptr(data))
	if errno != 0 {
		err := os.NewSyscallError("SYS_IOCTL", errno)
		return err
	}
	return nil
}