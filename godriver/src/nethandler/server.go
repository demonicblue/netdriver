package nethandler

import (
	"fmt"
	"ivbs"
	"nbd"
	"os"
	"runtime/debug"
	"syscall"
)

func passPacket(session *IVBSSession, resp *ivbs.Packet) {

	var reply *nbd.Reply

	entry, ok := session.Mapping[resp.Sequence]
	if !ok {
		fmt.Print("Not found in mapping: ")
		resp.Debug()
		return
	}

	reqRef := entry.Request

	switch resp.Op {
	case ivbs.OP_READ:
		reply = nbd.NewReply(reqRef, resp.DataSlice[ivbs.LEN_HEADER_PACKET:])

	case ivbs.OP_WRITE:
		reply = nbd.NewReply(reqRef, nil)

	default:
	}

	WriteBytesliceToFile(session.FdNetd, reply.Byteslice())

	delete(session.Mapping, resp.Sequence)
}

func serverListener(session *IVBSSession) {

	for !session.Quit {
		select {
		case response, chOk := <-session.ResponseCh:
			if !chOk {
				return
			}
			passPacket(session, response)
		}
	}
}

// Server thread
func server(session *IVBSSession) {

	/*tmp_file, err := os.OpenFile(session.NbdPath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Could not open device for testing.")
	}
	tmp_file.Close()*/

	// Panic recovering mechanism
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in server", r)
			debug.PrintStack()
			session.Quit = true
		}
	}()

	defer session.FdNetd.Close()
	defer syscall.Close(session.Fd[1])

	go serverListener(session)

	for !session.Quit {
		b := make([]byte, nbd.LEN_REQUEST_HEADER)
		var packet *ivbs.Packet

		_, err := ReadBytesliceFromFile(session.FdNetd, nbd.LEN_REQUEST_HEADER, b)
		if err != nil {
			fmt.Printf("Error reading in server.go:server():%s\n", err)
			session.Quit = true
			continue
		}

		request := nbd.NewRequest(b)

		switch request.Cmd {
		case nbd.NBD_CMD_READ:
			packet = ivbs.NewRead(session, request.From, uint64(request.Len))

		case nbd.NBD_CMD_WRITE:
			if request.Len > 0 {
				ReadBytesliceFromFile(session.FdNetd, len(request.Data), request.Data)
			}
			packet = ivbs.NewWrite(session, request.From, request.Len, request.Data)

		default:
			// Unknown
			fmt.Println("Unknown nbd request")
			continue
		}

		// Save request for later reference
		session.Mapping[packet.Sequence] = RequestMapping{packet, request}

		//fmt.Printf("Trying to write %d to ivbs\n", packet.DataLen+ivbs.LEN_HEADER_PACKET)
		n, _ := session.Conn.Write(packet.Byteslice())
		_ = n
		//fmt.Printf("Wrote %d bytes\n", n)

		// Disconnect on quit
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

func ReadBytesliceFromFile(f *os.File, l int, b []byte) ([]byte, error) {
	var n, n2 int
	var err error
	for ; n < l; n += n2 {
		n2, err = (*f).Read(b[n:l])
		if err != nil {
			break
		}
	}
	return b[:n], err
}

func WriteBytesliceToFile(f *os.File, b []byte) error {
	l := len(b)
	var n, n2 int
	var err error
	for ; n < l; n += n2 {
		n2, err = (*f).Write(b[n:l])
		if err != nil {
			break
		}
	}
	return err
}
