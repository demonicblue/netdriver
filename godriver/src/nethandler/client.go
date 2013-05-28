package nethandler

import (
	"fmt"
	"nbd"
	"syscall"
)

// Client thread
func client(session *IVBSSession) {
	
	// Close all files before quitting
	defer session.NbdFile.Close()
	defer syscall.Close(session.Fd[0])

	socket_fd := session.Fd[0]
	nbd_fd := session.NbdFile.Fd()
	
	fmt.Println("Starting client")
	if err:= nbd.CallUint64(nbd_fd, nbd.NBD_SET_SIZE, session.Size); err != nil {
		fmt.Printf("Error setting size: %s", err)
	}
	if err:= nbd.Call2(nbd_fd, nbd.NBD_CLEAR_SOCK, 0); err != nil {
		fmt.Printf("Error clearing socket: %s", err)
	}
	
	if err := nbd.Call2(nbd_fd, nbd.NBD_SET_SOCK, socket_fd); err != nil {
		fmt.Printf("Could not set socket: %s\n", err)
	}
	
	if err := nbd.Call2(nbd_fd, nbd.NBD_DO_IT, 0); err != nil {
		fmt.Printf("Error starting client: %s\n", err)
	}
	
	fmt.Println("Disconnecting..")
	
	//nbd.Call(nbd_fd, nbd.NBD_CLEAR_QUE, 0)
	//nbd.Call(nbd_fd, nbd.NBD_CLEAR_SOCK, 0)
	
}