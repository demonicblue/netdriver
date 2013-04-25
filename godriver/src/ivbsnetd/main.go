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
)

// Capitalized, hush hush
const (
	SERVER_SOCKET = 0
	CLIENT_SOCKET = 1
)

// Values imported from nbd.h
const (
	NBD_SET_SOCK	= 0
	NBD_SET_BLKSIZE = 1
	NBD_SET_SIZE	= 2
	NBD_DO_IT		= 3
	NBD_CLEAR_SOCK	= 4
	NBD_CLEAR_QUE	= 5
	NBD_PRINT_DEBUG	= 6
	NBD_SET_SIZE_BLOCKS	= 7
	NBD_DISCONNECT	= 8
	NBD_SET_TIMEOUT	= 9
	NBD_SET_FLAGS	= 10
)

const DATASIZE = 1024*1024*50

func nbd_request(request int) int {
	return int(C.nbd_request(C.int(request)))
}

func ioctl(a1, a2, a3 int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(a1), uintptr(a2), uintptr(a3))
	return err
}

func client(socket_fd, nbd_fd int) {
	
	if err := ioctl(nbd_fd, nbd_request(NBD_SET_SOCK), socket_fd); err != nil {
		fmt.Println(err)
	}
	
	if err := ioctl(nbd_fd, nbd_request(NBD_DO_IT), 0); err != nil {
		fmt.Println(err)
	}
	
	ioctl(nbd_fd, nbd_request(NBD_CLEAR_QUE), 0)
	ioctl(nbd_fd, nbd_request(NBD_CLEAR_SOCK), 0)
	
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
	
	go client(fd[CLIENT_SOCKET], nbd_fd)
	
	syscall.Close(fd[0])
	syscall.Close(fd[1])
	syscall.Close(nbd_fd)
	
	fmt.Println("Ending")
	
	
}

