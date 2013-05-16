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
	"strconv"
)

// Capitalized, hush hush
const (
	SERVER_SOCKET = 0
	CLIENT_SOCKET = 1
)
var httpAlive = make(chan int)
var lista map[int]string
var listm map[string]string

const DATASIZE = 1024*1024*50

const (
	NETD_DISC	= iota
)

// Switch byte-order
func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

func HttpCheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://reddit.com/r/golang.json") //insert json-object here
	if err != nil{
		fmt.Println("Error: %g", err)
	}
	if resp.StatusCode != http.StatusOK{
		fmt.Println(resp.Status)
	}
	fmt.Fprintf(w, "<h1>Health Status</h1>\nStatus: %s", resp.Status)
}

func HttpRootHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error:%g", err)
	}

	cmd := r.Form["command"][0]

	switch cmd{

		case "exit":
			fmt.Fprint(w, "<h1>The HTTP-Server is shutting down...</h1>")
			httpAlive <- 1
			break

		case "mount":
			//TODO Real mounting to NBD-devices with real images
			targetNBD := r.Form["nbd"][0]
			targetImg := r.Form["target"][0]
			for i:=0; i<len(lista); i++{
				if lista[i] == targetNBD{
					listm[lista[i]] = targetImg
					lista[i] = ""
					return
				}
			}
			for key, value := range lista{
				if value != ""{
					listm[lista[key]] = targetImg
					lista[key] = ""
					break
				}
			}
			break

		case "unmount":
			//TODO Real unmounting of NBD-devices
			targetNBD := r.Form["nbd"][0]
			for key, _ := range lista {
				if lista[key] == ""{
					delete(listm, targetNBD)
					lista[key] = targetNBD
					break
				}
			}
			break

		case "lista":
			for i:=0; i<len(lista); i++{
				if lista[i] != ""{
					fmt.Fprintln(w, lista[i])
				}
			}
			break

		case "listm":
			for key, value := range listm{
				fmt.Fprintln(w, key+"\t"+value)
			}
			break

	}
	return
}

// Client thread
func client(nbd_fd uintptr, socket_fd int) {
	
	fmt.Println("Starting client")
	if err:= nbd.Call2(nbd_fd, nbd.NBD_SET_SIZE, DATASIZE); err != nil {
		fmt.Printf("Error setting size: %g", err)
	}
	if err:= nbd.Call2(nbd_fd, nbd.NBD_CLEAR_SOCK, 0); err != nil {
		fmt.Print("Error clearing socket: %g", err)
	}
	
	if err := nbd.Call2(nbd_fd, nbd.NBD_SET_SOCK, socket_fd); err != nil {
		fmt.Printf("Could not set socket: %g\n", err)
	}
	
	if err := nbd.Call2(nbd_fd, nbd.NBD_DO_IT, 0); err != nil {
		fmt.Print("Error starting client: %g\n", err)
	}
	
	fmt.Println("Disconnecting..")
	
	//nbd.Call(nbd_fd, nbd.NBD_CLEAR_QUE, 0)
	//nbd.Call(nbd_fd, nbd.NBD_CLEAR_SOCK, 0)
	
}

const firstIVBSProxy string = "127.0.0.1:3033"

const MAX_CH_BUFF = 20

type IVBSSession struct {
	Conn net.Conn
	Id [32]byte
	Image string
	Username string
	Passwd string
	Send chan []byte
	Response chan IVBSResponse
	QuitCh chan bool
	NbdFile *os.File
	NbdPath string
}

type IVBSResponse struct {
	packet *ivbs.IvbsPacket
	data []byte
}

type IVBSRequest struct {
	Sequence uint32
	Handle [8]byte
	Type uint32
}

func sendPacket(session IVBSSession, op uint32, data []byte) {
	
}

func IOHandler(session IVBSSession) {
	if session.Conn == nil {
		//TODO Setup new connection
	}
	
	// Sender - receives data on channel and writes to connection
	go func(session IVBSSession) { //TODO Eliminate and refactor to server thread or improve
		data := <- session.Send
		session.Conn.Write(data)
	}(session)
	
	
	data := make([]byte, 45)// Header packet
	var moreData []byte 	//TODO Make before loop to save resources?
	quitIO := false
	
	
	for !quitIO {
		session.Conn.SetReadDeadline(time.Now().Add(2*time.Second)) // Make sure net.Read() doesn't block indefinetley
		_, err := session.Conn.Read(data)
		
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			// Received timeout, carry one
		} else if err != nil {
			// Fatal
		} else {
		
			reply := ivbs.IvbsSliceToStruct(data)
			if reply.DataLength > 0 {
				// Read more data
				moreData := make([]byte, reply.DataLength)
				session.Conn.Read(moreData)
			} else { 
				moreData = nil
			}
			
			switch reply.Op { // TODO Maybe handle greetings with reconnects
			case ivbs.OP_LIST_PROXIES:
			case ivbs.OP_READ, ivbs.OP_WRITE:
				session.Response <- IVBSResponse{reply, moreData}
			case ivbs.OP_KEEPALIVE:
			default:
				//Unknown
			}
		}
		select {
		case <- session.QuitCh:
			quitIO = true
		}
	}
	
}

func setupConnection(image, user, passwd, nbd_path string) (IVBSSession, error) {
	fmt.Println("Setting up connection to 127.0.0.1:3033")
	// Set up connection to IVBS
	conn, err := net.Dial("tcp", firstIVBSProxy)
	if err != nil {
		fmt.Printf("Connection failed: %g\n", err)
		//return
	}
	
	ivbs_slice := make([]byte, 45)
	conn.Read(ivbs_slice)
	ivbs_reply := ivbs.IvbsSliceToStruct(ivbs_slice)
	
	if ivbs_reply.Op != ivbs.OP_GREETING {
		fmt.Println("Error, received: %d", ivbs_reply.Op)
		//return
	}
	
	packet := new(ivbs.IvbsPacket)
	packet.Op = ivbs.OP_LOGIN
	packet.DataLength = ivbs.LEN_USERNAME + ivbs.LEN_PASSWORD_HASH
	
	loginPacket := new(ivbs.IvbsLogin)
	loginPacket.Name = user
	loginPacket.PasswordHash = passwd // TODO Hash that thing
	
	dataSlice := ivbs.IvbsStructToSlice(packet)
	dataSlice = append(dataSlice, ivbs.LoginStructToSlice(loginPacket)...)
	
	conn.Write(dataSlice)
	// Get reply
	conn.Read(ivbs_slice) // TODO Make sure reply is OK
	
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0) // Inter-process, client-server communication
	if(err != nil) {
		fmt.Printf("socketpair() failed with error: %g", err)
	}
	_ = fd
	
	nbd_file, err := os.OpenFile(nbd_path, os.O_RDWR, 0666)
	if(err != nil) {
		fmt.Printf("Tried opening %s with error: %g\nExiting..\n", nbd_path, err)
		os.Exit(0)
	}
	
	session := IVBSSession{
							conn, ivbs_reply.SessionId,
							image, user, passwd,
							make(chan []byte),
							make(chan IVBSResponse, MAX_CH_BUFF),
							make(chan bool),
							nbd_file,
							"",
	}
	
	// Start serving network data and return the session
	go IOHandler(session)
	return session, nil
	
	
}

// Server thread
func server(session IVBSSession) {
	/*request := new(nbd.Nbd_request)
	reply := new(nbd.Nbd_reply)
	_ = reply
	_ = request*/
	
	//nbd_fd := session.NbdFile.Fd()
	//defer nbd_file.Close()
	
	
	
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
		/*select {
		case <-quitCh:
			fmt.Println("Trying to disconnect..")
			return
			nbd.Call2(nbd_fd, nbd.NBD_CLEAR_QUE, 0)
			nbd.Call2(nbd_fd, nbd.NBD_DISCONNECT, 0)
			nbd.Call2(nbd_fd, nbd.NBD_CLEAR_SOCK, 0)
			syscall.Close(socket_fd)
			fmt.Println("Tried disconnecting..")
			return
		default:
			fmt.Println("Waiting..")
			time.Sleep(1000 * time.Millisecond)
		}*/
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

	var nbd_path, server string
	var nrDevices int
	
	// Setup flags
	flag.StringVar(&nbd_path, "n", "/dev/nbd5", "Path to NBD device")
	flag.StringVar(&server, "c", "localhost:12345", "Address for the HTTP-Server")
	flag.IntVar(&nrDevices, "d", 50, "Number of NBD-devices")
	flag.Parse()
	
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0) // Inter-process, client-server communication
	if(err != nil) {
		fmt.Printf("socketpair() failed with error: %g", err)
	}
	
	nbd_file, err := os.OpenFile(nbd_path, os.O_RDWR, 0666)
	if(err != nil) {
		fmt.Printf("Tried opening %s with error: %g\nExiting..\n", nbd_path, err)
		os.Exit(0)
	}
	_ = nbd_file.Fd()
	
	// Dat thread
	//go client(nbd_fd, fd[CLIENT_SOCKET])
	//fmt.Println("Server..")
	//go server(fd[SERVER_SOCKET], quitCh, nbd_path, nbd_file)
	
	fmt.Println("HTTP-Server starting on", server)

	lista = make(map[int]string)
	listm = make(map[string]string)

	for i:=0; i<nrDevices; i++{
		lista[i] = ("/dev/nbd"+strconv.Itoa(i))
	}
	
	http.HandleFunc("/", HttpRootHandler)
	http.HandleFunc("/check-health", HttpCheckHealthHandler)

	go http.ListenAndServe(server, nil)

	fmt.Println("HTTP-Server is up and running!")

	<-httpAlive

	fmt.Println("HTTP-Server shutting down...")

	time.Sleep(5 * time.Second)
	
	//disconnect(nbd_path, nbd_fd)
	
	time.Sleep(5 * time.Second)
	fmt.Println("Starting ending of main")
	syscall.Close(fd[0])
	syscall.Close(fd[1])
	//syscall.Close(nbd_fd)
	
	fmt.Println("Ending main")
}

