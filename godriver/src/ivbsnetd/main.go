package main

/*
//#include "nbd.h"
//#include <linux/types.h>
#include <sys/ioctl.h>
// Gets ioctl numbers for nbd commands
static int nbd_request(int cmd) {
	return _IO(0xab, cmd);
}

*/
import "C"


import(
	"fmt"
	"flag"
	"syscall"
	"os"
	"unsafe"
)

// Capitalized, hush hush
const (
	SERVER_SOCKET = 0
	CLIENT_SOCKET = 1
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

const DATASIZE = 1024*1024*50

type nbd_request struct {
	magic uint32
	cmd uint32
	handle [8]byte
	from uint64
	len uint32
}

type nbd_reply struct {
	magic uint32
	error uint32
	handle [8]byte
}

func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

func request(request int) int {
	return int(C.nbd_request(C.int(request)))
}

func ioctl(a1, a2, a3 int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(a1), uintptr(a2), uintptr(a3))
	return err
}

func client(socket_fd, nbd_fd int) {
	
	if err := ioctl(nbd_fd, request(NBD_SET_SOCK), socket_fd); err != nil {
		fmt.Println(err)
	}
	
	if err := ioctl(nbd_fd, request(NBD_DO_IT), 0); err != nil {
		fmt.Println(err)
	}
	
	ioctl(nbd_fd, request(NBD_CLEAR_QUE), 0)
	ioctl(nbd_fd, request(NBD_CLEAR_SOCK), 0)
	
}

func server(socket_fd int) {
	request := new(nbd_request)
	reply := new(nbd_reply)
	_ = reply
	_ = request
	b := make([]byte, unsafe.Sizeof(request))
	
	//fmt.Println(unsafe.Sizeof(*request))
	
	for {
		_, _ = syscall.Read(socket_fd, b)
		//copy(reply.handle, request.handle)
		
		len := ntohl(request.len)
		_ = len
		
		break
	}
}

func main() {
	data := make([]uint8, DATASIZE)
	_ = data[0] // TODO Remove
	
	var nbd_path string
	
	flag.StringVar(&nbd_path, "n", "DevicePath", "Path to NBD device")
	flag.Parse()
	
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if(err != nil) {
		fmt.Println(err)
	}
	
	nbd_fd, err := syscall.Open(nbd_path, syscall.O_RDWR, 0666)
	if(err != nil) {
		fmt.Printf("Tried opening %s with error: ", nbd_path)
		fmt.Println(err)
		fmt.Println("Exiting..")
		os.Exit(0)
	}
	
	ioctl(nbd_fd, request(NBD_SET_SIZE), DATASIZE)
	ioctl(nbd_fd, request(NBD_CLEAR_SOCK), 0)
	
	// Dat thread
	go server(nbd_fd)
	go client(fd[CLIENT_SOCKET], nbd_fd)
	
	
	
	syscall.Close(fd[0])
	syscall.Close(fd[1])
	syscall.Close(nbd_fd)
	
	fmt.Println("Ending")
	
	
}

