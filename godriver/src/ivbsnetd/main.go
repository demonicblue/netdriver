package main

import(
	"fmt"
	"flag"
	//"syscall"
	//"os"
	"time"
	//"nbd"
	//"net"
	//"ivbs"
	//"net/http"
	//"strconv"
	//"strings"
	"httpserver"
	"config"
)

// Capitalized, hush hush
const (
	SERVER_SOCKET = 0
	CLIENT_SOCKET = 1
)

const (
	NETD_DISC	= iota
)

var lista map[int]string
var listm map[string]string

// Switch byte-order
func ntohl(v uint32) uint32 {
	return uint32(byte(v >> 24)) | uint32(byte(v >> 16))<<8 | uint32(byte(v >> 8))<<16 | uint32(byte(v))<<24
}

func main() {

	var nbd_path, server string
	var nrDevices int
	var testUser, testPasswd string
	
	// Setup flags
	flag.StringVar(&testUser, "u", "", "Username")
	flag.StringVar(&testPasswd, "p", "", "Password")
	
	flag.StringVar(&nbd_path, "n", "/dev/nbd5", "Path to NBD device")
	flag.StringVar(&server, "c", ":8080", "Address for the HTTP-Server")
	flag.IntVar(&nrDevices, "d", 50, "Number of NBD-devices")
	flag.Parse()
	

	// Dat thread
	//go client(nbd_fd, fd[CLIENT_SOCKET])
	//fmt.Println("Server..")
	//go server(fd[SERVER_SOCKET], quitCh, nbd_path, nbd_file)
	//setupConnection("exjobb-test", testUser, testPasswd, nbd_path)
	
	httpAlive := httpserver.SetupHttp(server, nrDevices)

	config.ReadFile()

	<-httpAlive

	fmt.Println("HTTP-Server shutting down...")

	time.Sleep(5 * time.Second)
	
	//disconnect(nbd_path, nbd_fd)
	
	time.Sleep(5 * time.Second)
	fmt.Println("Starting ending of main")
	//syscall.Close(fd[0])
	//syscall.Close(fd[1])
	//syscall.Close(nbd_fd)
	
	fmt.Println("Ending main")
}
