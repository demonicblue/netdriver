package main

/*
//import "nbd.h"
//import <linux/types.h>
#include <sys/ioctl.h>
static int nbd_request(int cmd) {
return _IO(0xab, cmd);
}
*/
import "C"

import (
	"fmt"
	"flag"
	"syscall"
	"unsafe"
	"os"
	"ioctl"
	"time"
)



/*
 * This is the packet used for communication between client and
 * server. All data are in network byte order.
 */
type nbd_request struct {
	Magic uint32
	Type uint32 
	Handle [8]string
	From uint64
	Len uint32
}

/*
 * This is the reply packet that nbd-server sends back to the client after
 * it has completed an I/O request (or an error occurs).
 */
type nbd_reply struct {
	Magic uint32
	Error uint32
	Handle [8]string
}

/*
 * Constansts, defining NBD-operations and size of the NBD-device.
 * Some constants are fetched from nbd.h
 */
const (
	NBD_SET_SOCK = iota
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
	NBD_CMD_READ = iota
	NBD_CMD_WRITE
	NBD_CMD_DISC
	NBD_CMD_FLUSH
	NBD_CMD_TRIM
)

const (
	DATASIZE = 1024*1024*50

	SERVER_SOCK = 0
	CLIENT_SOCK = 1

	NBD_REQUEST_MAGIC = 0x25609513
	NBD_REPLY_MAGIC = 0x67446698
)

/*
 * NTOHL-function
 */
func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

/*
 * NBD Client-function
 */
func nbdclient(socket_fd, nbd_fd int) {
	if err := ioctl.Ioctl(nbd_fd, NBD_SET_SOCK, socket_fd); err != nil {
		fmt.Println("IOCTL_SET_SOCK had an error:", err)
	}
	if err := ioctl.Ioctl(nbd_fd, NBD_DO_IT, 0); err != nil {
		fmt.Println("IOCTL_DO_IT had an error:", err)
	}

	ioctl.Ioctl(nbd_fd, NBD_CLEAR_QUE, 0)
	ioctl.Ioctl(nbd_fd, NBD_CLEAR_SOCK, 0)
}

/*
 * Main-function.
 */
func main() {
	request := nbd_request{}
	reply := nbd_reply{}
	//void *data, *chunk

	//data := make([]byte, DATASIZE)
	var dev_path string
	flag.StringVar(&dev_path, "n", "/dev/nbd0", "Path to NBD device.")
	flag.Parse()

	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)

	nbd_fd, err := syscall.Open(dev_path, syscall.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
	}

	if err := ioctl.Ioctl(nbd_fd, NBD_SET_SIZE, DATASIZE); err != nil {
		fmt.Println("IOCTL_SET_SIZE had an error:", err)
	}

	if err := ioctl.Ioctl(nbd_fd, NBD_CLEAR_SOCK, 0); err != nil {
		fmt.Println("IOCTL_CLEAR_SOCK had an error:", err)
	}

	go nbdclient (fd[CLIENT_SOCK], nbd_fd)
	
	socket_fd := fd[SERVER_SOCK]

	reply.Magic = ntohl(NBD_REPLY_MAGIC)
	reply.Error = ntohl(0)

	for {
		p := []byte{}
		bytes_read, err := syscall.Read(socket_fd, p)

		if err != nil {
			fmt.Println("Error occurred whilst reading!", bytes_read)
		}

		//memcpy(reply.handle, request.handle, cap(request.handle))

		length := ntohl(request.Len)
		//offset := request.From


		/*if request.Magic != ntohl(NBD_REQUEST_MAGIC) {
			fmt.Println(1, "Data integrity check failed")
		}*/

		rq := ntohl(request.Type)

		switch rq{
			case NBD_CMD_READ:
				chunk := make([]byte, (length + uint32(unsafe.Sizeof(reply))))
				syscall.Write(socket_fd, chunk)
			
			case NBD_CMD_WRITE:
				chunk := make([]byte, length)
				syscall.Read(socket_fd, chunk)
				syscall.Write(socket_fd, chunk)
			
			case NBD_CMD_DISC:
				os.Exit(0)

			case NBD_CMD_FLUSH:

			case NBD_CMD_TRIM:

			default:
				time.Sleep(1000)
				fmt.Println("Unexpected NBD command: %d", rq)
		}
	}
}