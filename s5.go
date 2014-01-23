package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

const RECV_BUF_LEN = 1024

func handleConnection(conn net.Conn) {
	buf := make([]byte, RECV_BUF_LEN)
	defer conn.Close()
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	conn.Write([]byte{5, 0}) //accept,no need to auth

	if _, err = io.ReadFull(conn, buf[:4]); err != nil {
		return
	}
	cmd := buf[1]
	addr_type := buf[3]

	var addr_ip net.IP
	var addr_str string
	if addr_type == 1 { //IPV4
		if _, err = io.ReadFull(conn, buf[:4]); err != nil {
			return
		}

		addr_ip = net.IP(buf)
		addr_str = addr_ip.String()
	} else if addr_type == 3 { //domain name
		if _, err = io.ReadFull(conn, buf[:1]); err != nil {
			return
		}
		len := buf[0]
		if _, err = io.ReadFull(conn, buf[:len]); err != nil {
			return
		}
		addr_str = string(buf[:len])
		ips, _ := net.LookupIP(addr_str)
		addr_ip = ips[0]
	}

	if _, err = io.ReadFull(conn, buf[:2]); err != nil {
		return
	}
	port := binary.BigEndian.Uint16(buf)
	fmt.Println(addr_str, port)
	//only support connect mode
	if cmd != 1 {
		conn.Write([]byte{5, 7, 0, 1})
		return
	}
	addr_str = fmt.Sprintf("%s:%d", addr_ip.String(), port)

	remote, err := net.Dial("tcp", addr_str)

	if err != nil {
		conn.Write([]byte{5, 5, 0, 1, 0, 0, 0, 0, 0, 0})
		fmt.Println(err.Error())
		return
	}

	conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	defer remote.Close()

	func_transfer := func(c1, c2 net.Conn) {
		buf := make([]byte, RECV_BUF_LEN)
		for {
			n, err := c1.Read(buf)
			if err != nil || n == 0 {
				c2.Close()
				return
			}
			c2.Write(buf[0:n])
		}
	}
	go func_transfer(conn, remote)
	func_transfer(remote, conn)
}

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	fmt.Println("open socks5 services at:", port)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic("error listening:" + err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn)

	}
}
