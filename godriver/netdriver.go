package ivbsnetd

import (
	"fmt"
	//"net"
	"syscall"
	"errors"
	//import "nbd.h"
	//"C"
)

const (
	DATASIZE = 1024*1024*50
	serverConn = 0
	clientConn = 1
)

func main(){
	lol := []string{"hej","hopp"}
	fmt.Println(lol)
	netdriver(12,lol)
}

func netdriver(argc int, argv []string) {
	var fd [2]int
	//struct nbdrequest request
	//struct nbdreply reply
	//void *data, *chunk
	var len, bytes_read uint32
	var offset uint64

	data := make([]byte, DATASIZE)

	dev_path := argv[1]

	//fmt.Println(nbd_fd, dev_path, fd, len, bytes_read, offset, data)

	syscall.Socketpair(AF_UNIX, SOCK_STREAM, 0, fd)

	nbd_fd := syscall.Open(dev_path, O_RDWR)
	if nbd_fd == nil{
		errm := errors.New("Couldn't open %s", dev_path)
		fmt.Print(errm)
	}

	syscall.Syscall(syscall.SYS_IOCTL, nbd_fd, NBD_SET_SIZE, DATASIZE)
	syscall.Syscall(syscall.SYS_IOCTL, nbd_fd, NBD_CLEAR_SOCK)

	if !fork() {
		syscall.Close(fd[serverConn])
		socket_fd := fd[clientConn]

		if syscall.Syscall(syscall.SYS_IOCTL, nbd_fd, NBD_SET_SOCKET, socket_fd) == -1 {
			errm := errors.New("Cannot set client socket.")
			fmt.Print(errm)
		}

		err := syscall.Syscall(syscall.SYS_IOCTL, nbd_fd, NBD_DO_IT)
		fmt.Fprintf(stderr, "nbd device terminated with code %d", err)
		if err = -1 {
			fmt.Fprintf(stderr, "%s\n", )
		}

		syscall.Syscall(syscall.SYS_IOCTL, nbd_fd, NBD_CLEAR_QUE)
		syscall.Syscall(syscall.SYS_IOCTL, nbd_fd, NBD_CLEAR_SOCK)

		exit(0)
	}	

	close(fd[clientConn])
	socket_fd = fd[serverConn]

	reply.magic = htonl(NBD_REPLY_MAGIC)
	reply.error = htonl(0)

	for {
		bytes_read := read(socket_fd, &request, cap(request))

		//memcpy(reply.handle, request.handle, cap(request.handle))

		len := len(ntohl(request)) 
	}
}