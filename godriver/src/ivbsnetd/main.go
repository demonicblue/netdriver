package main

import(
	"fmt"
	"flag"
	"syscall"
	"os"
	"time"
	"ioctl"
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

const (
	NETD_DISC	= iota
)

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

type ivbs_packet struct {
	session_id [32]byte
	op uint32
	status int8
	data_length uint32
	sequence uint32
}

// Switch byte-order
func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

// Client thread
func client(nbd_fd int, socket_fd int) {
	
	if err := ioctl.Ioctl(nbd_fd, NBD_SET_SOCK, socket_fd); err != nil {
		fmt.Printf("Could not set socket: %s\n", err)
	}
	
	if err := ioctl.Ioctl(nbd_fd, NBD_DO_IT, 0); err != nil {
		fmt.Print("Error starting client: %s\n", err)
	}
	
	fmt.Println("Disconnecting..")
	
	ioctl.Ioctl(nbd_fd, NBD_CLEAR_QUE, 0)
	ioctl.Ioctl(nbd_fd, NBD_CLEAR_SOCK, 0)
	
}

// Server thread
func server(socket_fd ,nbd_fd int, quit chan int) {
	request := new(nbd_request)
	reply := new(nbd_reply)
	_ = reply
	_ = request
	//b := make([]byte, unsafe.Sizeof(request)) //TODO: Set size of slice with constant instead of using the unsafe packet
	
	for {
		/*_, _ = syscall.Read(socket_fd, b)
		//copy(reply.handle, request.handle)
		
		len := ntohl(request.len)
		_ = len
		
		break*/
		select {
		case <-quit:
			fmt.Println("Trying to disconnect..")
			ioctl.Ioctl(nbd_fd, NBD_DISCONNECT, 0)
			return
		default:
			fmt.Println("Waiting..")
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func main() {
	data := make([]uint8, DATASIZE)
	_ = data[0] // TODO Remove
	
	var nbd_path string
	
	// Setup flags
	flag.StringVar(&nbd_path, "n", "nil", "Path to NBD device")
	flag.Parse()
	
	// Inter-process, client-server communication
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if(err != nil) {
		fmt.Printf("socketpair() failed with error: %s", err)
	}
	
	nbd_fd, err := syscall.Open(nbd_path, syscall.O_RDWR, 0666)
	if(err != nil) {
		fmt.Printf("Tried opening %s with error: %s\nExiting..\n", nbd_path, err)
		os.Exit(0)
	}
	
	if err:= ioctl.Ioctl(nbd_fd, NBD_SET_SIZE, DATASIZE); err != nil {
		fmt.Printf("Error setting size: %s", err)
	}
	if err:= ioctl.Ioctl(nbd_fd, NBD_CLEAR_SOCK, 0); err != nil {
		fmt.Print("Error clearing socket: %s", err)
	}
	
	//quitCh := make(chan int)
	
	if err := ioctl.Ioctl(nbd_fd, NBD_SET_SOCK, fd[CLIENT_SOCKET]); err != nil {
		fmt.Printf("Could not set socket: %s\n", err)
	}
	
	// Dat thread
	//go client(nbd_fd, fd[CLIENT_SOCKET])
	//go server(nbd_fd, fd[SERVER_SOCKET], quitCh)
	
	time.Sleep(5 * time.Second)
	
	//quitCh <- 0
	
	syscall.Close(fd[0])
	syscall.Close(fd[1])
	syscall.Close(nbd_fd)
	
	fmt.Println("Ending")
	
	
}

