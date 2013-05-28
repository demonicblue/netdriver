package nethandler

import (
	"fmt"
	"time"
	"ivbs"
	"net"
	"os"

)

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
		//_, err := session.Conn.Read(data)
		data, err := ReadBytesliceFromConn(session.Conn, ivbs.LEN_HEADER_PACKET, data)
		
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			// Received timeout, carry one
			fmt.Println("Received timeout!")

		} else if err != nil {
			// Fatal, maybe reconnect?
			fmt.Printf("Error: %s\n", err)
			os.Exit(0)

		} else {

			reply := ivbs.IvbsSliceToStruct(data)

			/*//fmt.Printf("Got packet, op: %d\n", reply.Op)
			fmt.Print("Received: ")
			reply.Debug() // Prints out the whole package for debugging*/

			if reply.DataLen > 0 {
				// Read more data
				reply.DataSlice = make([]byte, ivbs.LEN_HEADER_PACKET + reply.DataLen)
				copy(reply.DataSlice, data)

				/*n, _ := session.Conn.Read(reply.DataSlice[ivbs.LEN_HEADER_PACKET:])
				fmt.Printf("Read %d bytes of extra data from ivbs.\n", n)*/
				
				ReadBytesliceFromConn(session.Conn, int(reply.DataLen), reply.DataSlice[ivbs.LEN_HEADER_PACKET:])


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
			case ivbs.OP_LOGIN, ivbs.OP_ATTACH_TO_IMAGE:
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

func ReadBytesliceFromConn(c *net.TCPConn, l int, b []byte) ([]byte, error) {
	var n, n2 int
	var err error
	for ; n < l; n += n2 {
		n2, err = (*c).Read(b[n:l])
		if err != nil {
			break
		}
	}
	return b[:n], err
}