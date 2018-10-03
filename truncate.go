package main

import (
	"flag"
	"log"
	"os"
)

var (
	size int64
	name string
)

func main() {
	flag.Int64Var(&size, "s", 1024, "file size to create")
	flag.StringVar(&name, "n", "out.bin", "file name to create")
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	err = f.Truncate(size)
	if err != nil {
		return err
	}
	f.WriteString("Hello")
	return nil
}
