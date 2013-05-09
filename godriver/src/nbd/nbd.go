package nbd

import (
	"ioctl"
	"os"
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

type Nbd_request struct {
	magic uint32
	cmd uint32
	handle [8]byte
	from uint64
	len uint32
}

type Nbd_reply struct {
	magic uint32
	error uint32
	handle [8]byte
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