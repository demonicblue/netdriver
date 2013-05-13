package main

import(
	"fmt"
	"flag"
	"syscall"
	"os"
	"time"
	"nbd"
	"net"
	"ivbs"
	"net/http"
)

// Capitalized, hush hush
const (
	SERVER_SOCKET = 0
	CLIENT_SOCKET = 1
)
var kanal = make(chan int)

const DATASIZE = 1024*1024*50

const (
	NETD_DISC	= iota
)

// Switch byte-order
func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

func HttpCheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://reddit.com/r/golang.json") //insert ivbs-server ip
	if err != nil{
		fmt.Println(err)
	}
	if resp.StatusCode != http.StatusOK{
		fmt.Println(resp.Status)
	}
	fmt.Fprintf(w, "<h1>Health Status</h1>\nStatus: %s", resp.Status)
}

func HttpRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>HTTPRootHandler</h1>\n")
}

func HttpQuitServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>The HTTP-Server is shutting down...</h1>")
	kanal <- 1
}
func HttpMountImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	target := r.Form["user"][0]

	fmt.Fprint(w, "<h1>TODO Mount an IVBS-image</h1>\n ", target)
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
	fmt.Println("Setting up connection to 127.0.0.1")
	// Set up connection to IVBS
	conn, err := net.Dial("tcp", "127.0.0.1")
	if err != nil {
		fmt.Println("Connection failed")
	}
	
	packet := new(ivbs.IvbsPacket)
	packet.Op = ivbs.OP_LOGIN
	packet.DataLength = ivbs.LEN_USERNAME + ivbs.LEN_PASSWORD_HASH
	
	loginPacket := new(ivbs.IvbsLogin)
	loginPacket.Name = "foo"
	loginPacket.PasswordHash = "bar"
	
	dataSlice := ivbs.IvbsStructToSlice(packet)
	dataSlice = append(dataSlice, ivbs.LoginStructToSlice(loginPacket)...)
	
	
	_ = nbd_fd // TODO: Remove later
	//conn.Write(dataSlice)
	fmt.Println(conn)
	
	//quitCh := make(chan int)
	
	// Dat thread
	//go client(nbd_fd, fd[CLIENT_SOCKET])
	//fmt.Println("Server..")
	//go server(nbd_fd, fd[SERVER_SOCKET], quitCh, nbd_path)
	
	fmt.Println("HTTP-Server starting...")
	
	http.HandleFunc("/", HttpRootHandler)
	http.HandleFunc("/check-health", HttpCheckHealthHandler)
	http.HandleFunc("/quit", HttpQuitServer)
	http.HandleFunc("/mount", HttpMountImage)

	go http.ListenAndServe("localhost:1234", nil)

	fmt.Println("HTTP-Server is up and running!")

	<-kanal

	fmt.Println("HTTP-Server shutting down...")

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

