package main

import (
	"path"
	"runtime"

	"github.com/Ars2014/IT-CatlangBot/daemon"
	"github.com/joho/godotenv"
)

func main() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not find file location")
	}
	filename := path.Join(path.Dir(file), ".env")

	err := godotenv.Load(filename)
	if err != nil {
		panic(err)
	}
	err = daemon.Run()
	if err != nil {
		panic(err)
	}
}
