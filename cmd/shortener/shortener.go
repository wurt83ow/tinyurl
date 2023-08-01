package main

import (
	"fmt"
	"server"
	"storage"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}

	err := storage.Load()
	if err != nil {
		fmt.Println(err)
	}

}
