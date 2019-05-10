package main

import (
	"os"

	"github.com/ld86/nat/node"
)

func main() {
	node := node.New()
	if len(os.Args) > 1 {
		node.Bootstrap(os.Args[1:])
	}
	node.Serve()
}
