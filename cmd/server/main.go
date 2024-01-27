package main

import (
	"os"
	"transmuxer/internal/server"
)

func main() {
	server.App().Run(os.Args)
}
