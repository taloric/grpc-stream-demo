package main

import (
	"flag"
)

func main() {
	port := "38888"
	flag.StringVar(&port, "port", port, "The server port")
	flag.IntVar(&ServerDelay, "delay", ServerDelay, "The server delay, unit: ms")
	flag.Parse()
	server_start(port)
}
