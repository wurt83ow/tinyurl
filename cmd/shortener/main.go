package main

import (
	"github.com/wurt83ow/tinyurl/cmd/shortener/server"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
