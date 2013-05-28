package main

import (
	"net"
	"fmt"
	"ivbs"
	"encoding/binary"
	"math/rand"
)

type ivbsSession struct {
	Id [32]byte
}

func handleConnection(conn net.Conn) {
	session := new(ivbsSession)
	binary.BigEndian.PutUint64( session.Id[:], uint64(rand.Uint32() | rand.Uint32() << 32) )
	
	packet := new(ivbs.IvbsPacket)
	binary.BigEndian.PutUint32(packet.SessionId[:], 50042)
	packet.Op = ivbs.OP_GREETING
	
	dataSlice := ivbs.IvbsStructToSlice(packet)
	conn.Write(dataSlice)
	reply := make([]byte, 45)
	
	for {
		conn.Read(reply)
		packet := ivbs.IvbsSliceToStruct(reply)
		if packet.SessionId != session.Id {
			conn.Close()
			return
		}
		
		switch packet.Op {
		case ivbs.OP_LOGIN:
			extra := make([]byte, packet.DataLength)
			conn.Read(extra)
			login := ivbs.LoginSliceToStruct(extra)
			fmt.Printf("User: %s", login.Name)
		}
	}
	
	
}

func main() {
	
	
	ln, err := net.Listen("tcp", ":3033")
	if err != nil {
		fmt.Printf("Failed listening: %g\n", err)
	}
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Failed accepting new connection: %g\n", err)
		}
		
		go handleConnection(conn)
	}
}