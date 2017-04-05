package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/fujiwara/brigade"
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "644", "file mode")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("require dest filename")
	}
	m, err := strconv.ParseInt(mode, 8, 64)
	if err != nil {
		log.Fatalf("file mode %s is wrong", mode)
	}
	brigade.Run(flag.Args()[0], os.FileMode(m))
}
