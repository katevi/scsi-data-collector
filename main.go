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
	period           = 1 * time.Second
	collectedDataDir = "./data"
	allDevicesKey    = ""
	appName          = "scsi-data-collector"
)

func main() {
	// log to custom file
	logFileName := fmt.Sprintf("%s.log", appName)
	// open log file
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()
	appLog := log.New(logFile, appName, log.Lshortfile|log.LstdFlags)

	if err := os.Mkdir(collectedDataDir, os.ModePerm); err != nil {
		appLog.Printf("Failed to create dir %s", err)
	}

	for {
		time.Sleep(period)
		blockDevices, _ := devicelister.GetBlockDevices(allDevicesKey)
		for _, blockDevice := range blockDevices {
			go execCommand(appLog, blockDevice)
		}
	}
}

func execCommand(appLog *log.Logger, device devicelister.BlockDevice) {
	cmdArgs := fmt.Sprintf("sg_logs -a %s", device.Name)
	cmd := exec.Command("bash", "-c", cmdArgs)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		appLog.Printf("Cmd execution completed with err %s %v", err, errb.String())
	}
	deviceName := device.Name
	deviceName = strings.ReplaceAll(deviceName, "/", "_")
	filepath := path.Join(collectedDataDir, deviceName)

	dumpDatatoFile(filepath,
		[]byte(fmt.Sprint("cmd:\n", cmdArgs)),
		[]byte(fmt.Sprint("time:\n", time.Now().String())),
		[]byte(fmt.Sprint("out:\n", outb.String())),
		[]byte(fmt.Sprint("err:\n", errb.String())))
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
