package main

import (
	"server"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
