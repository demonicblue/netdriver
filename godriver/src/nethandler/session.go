package nethandler

import (
	"net"
	"ivbs"
	"os"
	"fmt"
	"nbd"
	"syscall"
	"time"
)

// 50 Mb size in bytes
const DATASIZE = 1024*1024*50

// Client thread
func client(session IVBSSession) {
	
	socket_fd := int(session.NbdFile.Fd())
	nbd_fd := session.NbdFile.Fd()
	
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

const firstIVBSProxy string = "10.46.1.128:11417"

const MAX_CH_BUFF = 20

type IVBSSession struct {
	Conn net.Conn
	seqeuence uint32
	Id []byte
	Image string
	Username string
	Passwd string
	SendCh chan []byte
	ResponseCh chan *ivbs.Packet
	QuitCh chan bool
	NbdFile *os.File
	NbdPath string
	Fd [2]int
}

type IVBSResponse struct {
	packet *ivbs.Packet
	data []byte
}

type IVBSRequest struct {
	Sequence uint32
	Handle [8]byte
	Type uint32
}

func (session *IVBSSession) GetSequence() uint32 {
	session.seqeuence++
	return session.seqeuence
}

func (session *IVBSSession) WriteSession(b []byte) {
	copy(b, session.Id)
}

func parseGreeting(packet *ivbs.Packet) {
	// TODO Create this
}

func IOHandler(session *IVBSSession) {
	if session.Conn == nil {
		//TODO Setup new connection
	}
	
	// Sender - receives data on channel and writes to connection
	/*go func(session IVBSSession) { //TODO Eliminate and refactor to server thread or improve
		data := <- session.SendCh
		session.Conn.Write(data)
	}(session)*/
	
	
	//var moreData []byte 	//TODO Make before loop to save resources?
	quitIO := false
	
	
	for !quitIO {
		session.Conn.SetReadDeadline(time.Now().Add(10*time.Second)) // Make sure net.Read() doesn't block indefinetley

		data := make([]byte, ivbs.LEN_HEADER_PACKET)// Header packets
		_, err := session.Conn.Read(data)
		
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			// Received timeout, carry one

		} else if err != nil {
			// Fatal, maybe reconnect?
			fmt.Printf("Error: %s\n", err)
			os.Exit(0)

		} else {

			fmt.Println("Got packet")

			reply := ivbs.IvbsSliceToStruct(data)

			if reply.DataLen > 0 {
				// Read more data
				reply.DataSlice = make([]byte, ivbs.LEN_HEADER_PACKET + reply.DataLen)
				copy(reply.DataSlice, data)
				session.Conn.Read(reply.DataSlice[ivbs.LEN_HEADER_PACKET:])

			} else {
				// Only header data
				reply.DataSlice = make([]byte, ivbs.LEN_HEADER_PACKET)
				copy(reply.DataSlice, data)

			}

			switch reply.Op { // TODO Maybe handle greetings with reconnects

			case ivbs.OP_LIST_PROXIES:
			case ivbs.OP_READ, ivbs.OP_WRITE:
				session.ResponseCh <- reply
			case ivbs.OP_KEEPALIVE:
			case ivbs.OP_GREETING:
				session.ResponseCh <- reply
			case ivbs.OP_LOGIN:
				session.ResponseCh <- reply
			default:
				//Unknown
			}
		}
		select {
		case <- session.QuitCh:
			quitIO = true
			fmt.Println("Received quit")
		default:
		}
	}
	
}

func SetupConnection(image, user, passwd, nbd_path string) (IVBSSession, error) {
	fmt.Println("Setting up connection to "+firstIVBSProxy)
	
	// Set up connection to IVBS
	conn, err := net.Dial("tcp", firstIVBSProxy)
	
	if nerr, ok := err.(net.Error); ok {
		fmt.Print(nerr.Error())
		return IVBSSession{}, err
	} else if err != nil {
		fmt.Printf("Connection failed: %g\n", err)
		return IVBSSession{}, err
	}
	
	session := IVBSSession{
							conn,
							0,
							make([]byte, ivbs.LEN_SESSIONID),
							image, user, passwd,
							make(chan []byte),
							make(chan *ivbs.Packet, MAX_CH_BUFF),
							make(chan bool),
							nil,
							nbd_path,
							[2]int{0, 0},
	}
	
	go IOHandler(&session)
	
	ivbs_reply :=<- session.ResponseCh
	
	if ivbs_reply.Op != ivbs.OP_GREETING {
		fmt.Println("Error, received: %d", ivbs_reply.Op)
		//return
	}
	
	copy(session.Id, ivbs_reply.SessionId) // Retreive session id from greeting packet

	// Create login packet
	packet := ivbs.NewLogin(&session, user, passwd)

	n, err := conn.Write(packet.Byteslice())
	if err != nil {
		fmt.Printf("Could not write login packet, wrote %d bytes with error: %s\n", n, err)
	} else {
		fmt.Printf("Wrote %d bytes login packet without error\n", n)
	}
	
	// Get reply
	ivbs_reply =<- session.ResponseCh // TODO Make sure reply is OK
	
	//fmt.Printf("Response seqeuence: %d, Op: %d, status: %d\n", ivbs_reply.Sequence, ivbs_reply.Op, ivbs_reply.Status)
	
	if ivbs_reply.Op != ivbs.OP_LOGIN || ivbs_reply.Status != ivbs.STATUS_OK {
		os.Exit(0)
	}

	fmt.Println("Logged in successfully!")



	
	
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0) // Inter-process, client-server communication
	if(err != nil) {
		fmt.Printf("socketpair() failed with error: %g", err)
	}
	session.Fd = fd
	
	nbd_file, err := os.OpenFile(nbd_path, os.O_RDWR, 0666)
	if(err != nil) {
		fmt.Printf("Tried opening %s with error: %g\nExiting..\n", nbd_path, err)
		os.Exit(0)
	}
	session.NbdFile = nbd_file
	
	// Start serving network data and return the session
	go client(session)
	
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