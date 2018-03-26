package main

import (
	"github.com/Ars2014/IT-CatlangBot/daemon"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	err = daemon.Run()
	if err != nil {
		panic(err)
	}
}
