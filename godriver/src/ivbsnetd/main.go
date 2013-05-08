package main

import(
	"fmt"
	"flag"
	"syscall"
	"os"
	"time"
	"nbd"
	"net"
	"encoding/binary"
)

// Capitalized, hush hush
const (
	SERVER_SOCKET = 0
	CLIENT_SOCKET = 1
)

const DATASIZE = 1024*1024*50

const (
	NETD_DISC	= iota
)

const (
	OP_GREETING		uint32 = 1
	OP_KEEPALIVE	uint32 = 2

	OP_LOGIN		uint32 = 100

	OP_READ			uint32 = 300
	OP_WRITE		uint32 = 301

	OP_LIST_PROXIES	uint32 = 211
)

const (
    STATUS_OK                = 0
    STATUS_NOT_FOUND         = 1
    STATUS_ALREADY_EXISTS    = 2
    STATUS_NOT_LOGGEDIN      = 3
    STATUS_PERMISSION_DENIED = 4
    STATUS_READ_ONLY         = 5
    STATUS_NOT_READY         = 6

    STATUS_TEMPORARY_ERROR   = 10
    STATUS_PERMANENT_ERROR   = 11
    STATUS_USE_ANOTHER_PROXY = 12
    STATUS_INVALID_REQUEST   = 13
    STATUS_EXPIRED           = 14
)

type Login struct {
    Name         string
    PasswordHash string
}
const LEN_USERNAME      = 32
const LEN_PASSWORD_HASH = 128

type ivbs_packet struct {
	sessionId [32]byte
	op uint32
	status int8
	dataLength uint32
	sequence uint32
}

// Switch byte-order
func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

// Client thread
func client(nbd_fd uintptr, socket_fd int) {
	
	fmt.Println("Starting client")
	if err:= nbd.Call2(nbd_fd, nbd.NBD_SET_SIZE, DATASIZE); err != nil {
		fmt.Printf("Error setting size: %s", err)
	}
	if err:= nbd.Call2(nbd_fd, nbd.NBD_CLEAR_SOCK, 0); err != nil {
		fmt.Print("Error clearing socket: %s", err)
	}
	
	if err := nbd.Call2(nbd_fd, nbd.NBD_SET_SOCK, socket_fd); err != nil {
		fmt.Printf("Could not set socket: %s\n", err)
	}
	
	if err := nbd.Call2(nbd_fd, nbd.NBD_DO_IT, 0); err != nil {
		fmt.Print("Error starting client: %s\n", err)
	}
	
	fmt.Println("Disconnecting..")
	
	//nbd.Call(nbd_fd, nbd.NBD_CLEAR_QUE, 0)
	//nbd.Call(nbd_fd, nbd.NBD_CLEAR_SOCK, 0)
	
}

// Server thread
func server(nbd_fd uintptr, socket_fd int, quitCh chan int, nbd_path string) {
	request := new(nbd.Nbd_request)
	reply := new(nbd.Nbd_reply)
	_ = reply
	_ = request
	//b := make([]byte, unsafe.Sizeof(request)) //TODO: Set size of slice with constant instead of using the unsafe packet
	
	time.Sleep(500*time.Millisecond)
	fmt.Println("In server: After sleep")
	/*tmp_file, err := os.OpenFile(nbd_path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Could not open device for testing.")
	}
	fmt.Println("In server: After open")
	tmp_file.Close()*/
	
	fmt.Println("Starting server loop..")
	for {
		/*_, _ = syscall.Read(socket_fd, b)
		//copy(reply.handle, request.handle)
		
		len := ntohl(request.len)
		_ = len
		
		break*/
		select {
		case <-quitCh:
			return
			fmt.Println("Trying to disconnect..")
			nbd.Call2(nbd_fd, nbd.NBD_CLEAR_QUE, 0)
			nbd.Call2(nbd_fd, nbd.NBD_DISCONNECT, 0)
			nbd.Call2(nbd_fd, nbd.NBD_CLEAR_SOCK, 0)
			syscall.Close(socket_fd)
			fmt.Println("Tried disconnecting..")
			return
		default:
			fmt.Println("Waiting..")
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func disconnect(nbd_path string, nbd_fd uintptr) {
	fmt.Println("Alternative disc..")
	tmp_fd := nbd_fd
	fmt.Println("Trying clear que")
	nbd.Call2(tmp_fd, nbd.NBD_CLEAR_QUE, 0)
	fmt.Println("Trying disc")
	nbd.Call2(tmp_fd, nbd.NBD_DISCONNECT, 0)
	fmt.Println("Trying clear sock")
	nbd.Call2(tmp_fd, nbd.NBD_CLEAR_SOCK, 0)
}

func ivbsStructToSlice(packet *ivbs_packet) ([]byte) {
	data := make([]byte, 45)
	
	copy(data[:32], packet.sessionId[:])
	binary.BigEndian.PutUint32(data[32:36], packet.op)
	data[36] = byte(packet.status)
	binary.BigEndian.PutUint32(data[37:41], packet.dataLength)
	binary.BigEndian.PutUint32(data[41:], packet.sequence)
	
	return data
}

func loginStructToSlice(packet *Login) ([]byte) {
	data := make([]byte, LEN_USERNAME + LEN_PASSWORD_HASH)
	
	copy(data[:LEN_USERNAME], []byte(packet.Name))
	copy(data[LEN_USERNAME:], []byte(packet.PasswordHash))
	
	return data
}

func main() {
	data := make([]uint8, DATASIZE)
	_ = data[0] // TODO Remove
	
	var nbd_path string
	
	// Setup flags
	flag.StringVar(&nbd_path, "n", "/dev/nbd5", "Path to NBD device")
	flag.Parse()
	
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0) // Inter-process, client-server communication
	if(err != nil) {
		fmt.Printf("socketpair() failed with error: %s", err)
	}
	
	//nbd_fd, err := syscall.Open(nbd_path, syscall.O_RDWR, 0666)
	nbd_file, err := os.OpenFile(nbd_path, os.O_RDWR, 0666)
	nbd_fd := nbd_file.Fd()
	defer nbd_file.Close()
	if(err != nil) {
		fmt.Printf("Tried opening %s with error: %s\nExiting..\n", nbd_path, err)
		os.Exit(0)
	}
	
	// Set up connection to IVBS
	conn, err := net.Dial("tcp", "10.0.0.1")
	if err != nil {
		fmt.Println("Connection failed")
	}
	
	packet := new(ivbs_packet)
	packet.op = OP_LOGIN
	packet.dataLength = LEN_USERNAME + LEN_PASSWORD_HASH
	
	loginPacket := new(Login)
	loginPacket.Name = "foo"
	loginPacket.PasswordHash = "bar"
	
	dataSlice := ivbsStructToSlice(packet)
	dataSlice = append(dataSlice, loginStructToSlice(loginPacket)...)
	
	
	_ = nbd_fd // TODO: Remove later
	conn.Write(dataSlice)
	
	//quitCh := make(chan int)
	
	// Dat thread
	//go client(nbd_fd, fd[CLIENT_SOCKET])
	//fmt.Println("Server..")
	//go server(nbd_fd, fd[SERVER_SOCKET], quitCh, nbd_path)
	
	time.Sleep(5 * time.Second)
	
	//quitCh <- 0
	
	//tmp_fd := nbd_fd
	//disconnect(nbd_path, nbd_fd)
	
	time.Sleep(5 * time.Second)
	fmt.Println("Starting ending of main")
	syscall.Close(fd[0])
	syscall.Close(fd[1])
	//syscall.Close(nbd_fd)
	
	fmt.Println("Ending main")
	
	
}

