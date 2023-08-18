package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	devicelister "scsicollector/internal"
	"strings"

	"time"
)

const (
	//period = 10 * time.Minutes
	period     = 1 * time.Second
	dir        = "./data"
	allDevices = ""
)

func main() {
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		log.Printf("Failed to create dir %s", err)
	}

	for {
		time.Sleep(period)
		blockDevices, _ := devicelister.GetBlockDevices(allDevices)
		for _, blockDevice := range blockDevices {
			go execCommand(blockDevice)
		}
	}
}

func execCommand(device devicelister.BlockDevice) {
	cmdArgs := fmt.Sprintf("sg_logs -a %s", device.Name)
	cmd := exec.Command("bash", "-c", cmdArgs)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Printf("Cmd execution completed with err %s %v", err, errb.String())
	}
	deviceName := device.Name
	deviceName = strings.ReplaceAll(deviceName, "/", "_")
	filepath := path.Join(dir, deviceName)

	dumpDatatoFile(filepath, []byte(cmdArgs),
		[]byte(time.Now().String()), outb.Bytes(), errb.Bytes())
}

func dumpDatatoFile(filepath string, contents ...[]byte) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, content := range contents {
		f.Write(content)
		f.Write([]byte("\n"))
	}
}
