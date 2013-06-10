package main

import (
	"flag"
	"fmt"
	//"time"
	"config"
	"httpserver"
)

var lista map[int]string
var listm map[string]string

// Switch byte-order
func ntohl(v uint32) uint32 {
	return uint32(byte(v>>24)) | uint32(byte(v>>16))<<8 | uint32(byte(v>>8))<<16 | uint32(byte(v))<<24
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
	/*
		go client(nbd_fd, fd[CLIENT_SOCKET])
		fmt.Println("Server..")
		go server(fd[SERVER_SOCKET], quitCh, nbd_path, nbd_file)
		setupConnection("exjobb-test", testUser, testPasswd, nbd_path)
	*/

	httpAlive := httpserver.SetupHttp(server, nrDevices)

	config.ReadFile()

	<-httpAlive

	fmt.Println("HTTP-Server shutting down...")

	//disconnect(nbd_path, nbd_fd)

	fmt.Println("Ending main")
}
