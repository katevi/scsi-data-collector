package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

const (
	//period = 10 * time.Minutes
	period = 1 * time.Second
	dir    = "./data"
)

func main() {
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		log.Printf("Failed to create dir %s", err)
	}

	for {
		time.Sleep(period)
		execCommand()
	}
}

func execCommand() {
	log.Printf("Hello, world!")
	cmd := exec.Command("bash", "-c", "sg_logs -a /dev/sda")

	// open the out file for writing
	t := time.Now().UTC().Format("2006-01-02-15-04-05")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Printf("Cmd execution completed with err %s %v", err, errb.String())
	}
	err = os.WriteFile(path.Join(dir, t), outb.Bytes(), os.ModePerm)
	if err != nil {
		log.Printf("Writing result to file completed with err %s %v", err, errb)
	}
}
