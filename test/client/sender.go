package main

import (
	"fmt"
	"log"
	"net"
)

var conn *net.TCPConn

func connectFs() error {
	server := "127.0.0.1:1024"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		log.Printf("Fatal error: %s", err.Error())
		return err
	}
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("Fatal error: %s", err.Error())
		return err
	}
	return nil
}

func sendToFs(words string) error {
	conn.Write([]byte(words))

	buffer := make([]byte, 2048)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(conn.RemoteAddr().String(), "waiting server back msg error: ", err)
		return fmt.Errorf("waiting server back msg error: %s", err)
	}
	log.Println(conn.RemoteAddr().String(), ": ", string(buffer[:n]))
	return nil
}

func closeConn() {
	conn.Close()
}
