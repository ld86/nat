package network

import (
	"log"
	"net"
	"os"
	"syscall"
)

// LocalIP returns output IP address
func LocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// CreatePacketConn creates UDP connection ready for input and output messages
func CreatePacketConn() net.PacketConn {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)

	if err != nil {
		log.Fatalf("Cannot create socket, %s", err)
	}

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatalf("Cannot set SO_REUSEADDR on socket, %s", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil && udpAddr.IP != nil {
		log.Fatalf("Cannot resolve addr, %s", err)
	}

	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Port: udpAddr.Port}); err != nil {
		log.Fatalf("Cannot bind socket, %s", err)
	}

	file := os.NewFile(uintptr(fd), string(fd))
	conn, err := net.FilePacketConn(file)
	if err != nil {
		log.Fatalf("Cannot create connection from socket, %s", err)
	}

	if err = file.Close(); err != nil {
		log.Fatalf("Cannot close dup file, %s", err)
	}

	return conn
}
