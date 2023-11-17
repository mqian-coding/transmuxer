package main

import (
	"concurrency-practice/internal/server"
	"os"
)

func main() {
	server.App().Run(os.Args)
}
