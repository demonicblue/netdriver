package ivbsnetd

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
	"syscall"
	"errors"
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
 *
 */
const (
	NBD_SET_SOCK = 0
	NBD_SET_BLKSIZE = 1
	NBD_SET_SIZE = 2
	NBD_DO_IT = 3
	NBD_CLEAR_SOCK = 4
	NBD_CLEAR_QUE = 5
	NBD_PRINT_DEBUG = 6
	NBD_SET_SIZE_BLOCKS = 7
	NBD_DISCONNECT = 8
	NBD_SET_TIMEOUT = 9
	NBD_SET_FLAGS = 10

	DATASIZE = 1024*1024*50

	SERVER_SOCK = 0
	CLIENT_SOCK = 1
)

func main(){
	//TODO
	lol := []string{"lol"}
	netdriver(1, lol)
}


func nbdrequest(request int) int {
	return int(C.nbd_request(C.int(request)))
}


func ioctl(a1, a2, a3 int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(a1), uintptr(a2), uintptr(a3))
	return err
}


func nbdclient(socket_fd, nbd_fd int) {
	if err := ioctl(nbd_fd, nbdrequest(NBD_SET_SOCK), socket_fd); err != nil {
		fmt.Println(err)
	}

	if err := ioctl(nbd_fd, nbdrequest(NBD_DO_IT), 0); err != nil {
		fmt.Println(err)
	}

	ioctl(nbd_fd, nbdrequest(NBD_CLEAR_QUE), 0)
	ioctl(nbd_fd, nbdrequest(NBD_CLEAR_SOCK), 0)
}


func netdriver(argc int, argv []string) {
	var fd [2]int
	request := nbd_request{}
	reply := nbd_reply{}
	//void *data, *chunk
	var len, bytes_read uint32

	data := make([]byte, DATASIZE)

	dev_path := argv[1]

	//fmt.Println(nbd_fd, dev_path, fd, len, bytes_read, offset, data)

	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)

	nbd_fd, err := syscall.Open(dev_path, syscall.O_RDWR, 0666)
	if nbd_fd == -1 {
		errm := errors.New("Couldn't open the device!")
		fmt.Print(errm)
	}

		ioctl(nbd_fd, nbdrequest(NBD_SET_SIZE), DATASIZE)
		ioctl(nbd_fd, nbdrequest(NBD_CLEAR_SOCK), 0)

	go nbdclient (fd[CLIENT_SOCK], nbd_fd)
		
	socket_fd := fd[SERVER_SOCK]

	reply.Magic = htonl(NBD_REPLY_MAGIC)
	reply.Error = htonl(0)

	for {
		bytes_read := read(socket_fd, &request, cap(request))

		//memcpy(reply.handle, request.handle, cap(request.handle))

		len := len(ntohl(request)) 

		offset := ntohl(request.From)
	}
}