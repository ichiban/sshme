package main

import (
	"flag"
	"os"

	"github.com/ichiban/sshme"
)

func main() {
	var s sshme.Server

	flag.StringVar(&s.Bind, "bind", os.Getenv("BIND"), "-bind :2022")
	flag.StringVar(&s.Key, "key", os.Getenv("KEY"), "-key ~/.ssh/sshme.rsa")
	flag.Parse()

	s.Run()
}
