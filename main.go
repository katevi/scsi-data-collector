package main

import (
	"log"
	"os/exec"
)

func main() {
	log.Printf("Hello, world!")
	cmd := exec.Command("sg_logs -a /dev/sda")
	err := cmd.Run()
	log.Printf("Cmd completed %s", err)
}
