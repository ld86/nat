package node

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/ld86/nat/network"
)

type Node struct {
	id         string
	addr       NodeAddr
	conn       net.PacketConn
	knownNodes map[string]NodeAddr
}

type NodeAddr struct {
	LocalIP    string
	LocalPort  string
	RemoteIP   string
	RemotePort string
}

type Message struct {
	SourceID   string
	SourceAddr NodeAddr
	KnownNodes map[string]NodeAddr
}

func New() *Node {
	conn := network.CreatePacketConn()
	localIP := network.LocalIP().String()
	localPort := strings.Split(conn.LocalAddr().String(), ":")[1]
	addr := NodeAddr{LocalIP: localIP, LocalPort: localPort}
	var id [20]byte
	rand.Read(id[:])

	return &Node{
		id:         hex.EncodeToString(id[:]),
		addr:       addr,
		conn:       conn,
		knownNodes: make(map[string]NodeAddr),
	}
}

func (node *Node) handleInboundMessages() {
	var buffer [1024]byte
	for {
		n, sourceRemoteAddr, err := node.conn.ReadFrom(buffer[:])
		if err != nil {
			log.Printf("Cannot ReadFrom conn, %s", err)
		}
		var message Message
		json.Unmarshal(buffer[:n], &message)
		fmt.Println(message)
		sourceAddr := node.knownNodes[message.SourceID]
		sourceAddr.LocalIP = message.SourceAddr.LocalIP
		sourceAddr.LocalPort = message.SourceAddr.LocalPort
		sourceAddr.RemoteIP = strings.Split(sourceRemoteAddr.String(), ":")[0]
		sourceAddr.RemotePort = strings.Split(sourceRemoteAddr.String(), ":")[1]
		node.knownNodes[message.SourceID] = sourceAddr
	}
}

func (node *Node) Ping(remoteIP string) {
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteIP)
	if err != nil {
		log.Printf("Cannot resolve %s, %s", remoteIP, err)
		return
	}
	message := Message{SourceID: node.id, SourceAddr: node.addr, KnownNodes: node.knownNodes}
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Cannot marshal message %s, %s", message, err)
	}
	node.conn.WriteTo(data, remoteAddr)

}

func (node *Node) Bootstrap(remoteIPs []string) {
	for _, remoteIP := range remoteIPs {
		node.Ping(remoteIP)
	}
}

func (node *Node) Serve() {
	fmt.Printf("%s:%s\n", node.addr.LocalIP, node.addr.LocalPort)
	go node.handleInboundMessages()
	for {
		time.Sleep(time.Second)
		for _, remoteAddr := range node.knownNodes {
			fmt.Printf("Sending my information to %s\n", remoteAddr)
			node.Ping(fmt.Sprintf("%s:%s", remoteAddr.LocalIP, remoteAddr.LocalPort))
			node.Ping(fmt.Sprintf("%s:%s", remoteAddr.RemoteIP, remoteAddr.RemotePort))
		}
	}
}
