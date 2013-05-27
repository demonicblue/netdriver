package nethandler

import (
	"fmt"
	"nbd"
	"syscall"
	"ivbs"
	//"os"
)

func passPacket(session *IVBSSession, response *ivbs.Packet) {

	var reply *nbd.Reply

	requestRef := session.Mapping[response.Sequence].Request
	//fd := session.Fd[1]
	//fd := session.FdNetd.Fd()

	switch response.Op {
	case ivbs.OP_READ:
		reply = nbd.NewReply(requestRef, response.DataSlice[ivbs.LEN_HEADER_PACKET:])
		fmt.Println("Sending read to nbd")
	case ivbs.OP_WRITE:
		reply = nbd.NewReply(requestRef, nil)
		fmt.Println("Sending write to nbd")
	default:
	}

	//syscall.Write(fd, reply.Byteslice())
	session.FdNetd.Write(reply.Byteslice())
}

func serverListener(session *IVBSSession) {
	//var packet *ivbs.Packet
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
	/*request := new(nbd.Nbd_request)
	reply := new(nbd.Nbd_reply)
	_ = reply
	_ = request*/
	
	//nbd_fd := session.NbdFile.Fd()
	//defer nbd_file.Close()

	/*tmp_file, err := os.OpenFile(session.NbdPath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Could not open device for testing.")
	}
	fmt.Println("In server: After open")
	tmp_file.Close()*/
	defer session.FdNetd.Close()
	defer syscall.Close(session.Fd[1])

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in server", r)
			session.Quit = true
		}
	}()

	go serverListener(session)

	//fd := session.Fd[1]
	//syscall.SetNonblock(fd, true)
	
	fmt.Println("Starting server loop..")
	for !session.Quit {
		b := make([]byte, nbd.LEN_REQUEST_HEADER)
		var packet *ivbs.Packet

		//_, _ = syscall.Read(fd, b)
		_, err := session.FdNetd.Read(b)
		if err != nil {
			fmt.Println("Error reading in server.go:server()")
			session.Quit = true
			continue
		}

		request := nbd.NewRequest(b)

		switch request.Cmd {
		case nbd.NBD_CMD_READ:
			packet = ivbs.NewRead(session, request.From, uint64(request.Len))
			session.Conn.Write(packet.Byteslice())
			fmt.Println("Sent read")
		case nbd.NBD_CMD_WRITE:
			if request.Len > 0 {
				//syscall.Read(fd, request.Data)
				session.FdNetd.Read(request.Data)
			}
			packet = ivbs.NewWrite(session, request.From, request.Len, request.Data)
			packet.Debug()
			session.Conn.Write(packet.Byteslice())
			fmt.Println("Sent write")
		default:
			// Unknown
			fmt.Println("Unknown nbd request")
		}

		// Save request for later reference
		session.Mapping[packet.Sequence] = RequestMapping{packet, request}
		//packet.Debug()



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