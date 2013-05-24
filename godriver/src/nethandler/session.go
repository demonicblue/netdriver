package nethandler

import (
	"net"
	"ivbs"
	"os"
	"fmt"
	"nbd"
	"syscall"
)

// 50 Mb size in bytes
const DATASIZE = 1024*1024*50

const firstIVBSProxy string = "10.46.1.128:11417"

const MAX_CH_BUFF = 20

type IVBSSession struct {
	Conn net.Conn
	seqeuence uint32
	Id []byte
	Image string
	Size uint64
	Username string
	Passwd string
	SendCh chan []byte
	ResponseCh chan *ivbs.Packet
	QuitCh chan bool
	Quit bool
	NbdFile *os.File
	NbdPath string
	Fd [2]int
	Mapping map[uint32]RequestMapping
}

/*
type IVBSResponse struct {
	packet *ivbs.Packet
	data []byte
}

type IVBSRequest struct {
	Sequence uint32
	Handle [8]byte
	Type uint32
}
*/
type RequestMapping struct {
	Packet *ivbs.Packet
	Request *nbd.Request
}

func (session *IVBSSession) GetSequence() uint32 {
	session.seqeuence++
	return session.seqeuence
}

func (session *IVBSSession) WriteSession(b []byte) {
	copy(b, session.Id)
}

func parseGreeting(session *IVBSSession, packet *ivbs.Packet) {
	copy(session.Id, packet.SessionId)

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
							image, 0, user, passwd,
							make(chan []byte),
							make(chan *ivbs.Packet, MAX_CH_BUFF),
							make(chan bool),
							false,
							nil,
							nbd_path,
							[2]int{0, 0},
							make(map[uint32] RequestMapping),
	}
	
	go IOHandler(&session)
	
	ivbs_reply :=<- session.ResponseCh
	
	if ivbs_reply.Op != ivbs.OP_GREETING {
		fmt.Println("Error, received: %d", ivbs_reply.Op)
		//return
	}
	
	parseGreeting(&session, ivbs_reply)

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

	packet = ivbs.NewAttach(&session, image)

	n, err = conn.Write(packet.Byteslice())
	if err != nil {
		fmt.Printf("Could not write attach packet, wrote %d bytes with error: %s\n", n, err)
	} else {
		fmt.Printf("Wrote %d bytes attach packet without error\n", n)
	}

	ivbs_reply =<- session.ResponseCh

	if ivbs_reply.Op != ivbs.OP_ATTACH_TO_IMAGE || ivbs_reply.Status != ivbs.STATUS_OK {
		fmt.Printf("Received error in attach packet. OP: %d, Status: %d\n", ivbs_reply.Op, ivbs_reply.Status)
		os.Exit(0)
	}

	//TODO Save image size
	attach := ivbs.AttachFromSlice(ivbs_reply)
	session.Size = attach.Size
	fmt.Printf("Image size: %d byte, %d MB\n", session.Size, session.Size/1024/1024)

	
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
	go client(&session)
	go server(&session)

	tmp_file, err := os.OpenFile(session.NbdPath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Could not open device for testing.")
	}
	fmt.Println("In server: After open")
	tmp_file.Close()
	
	return session, nil
	
	
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