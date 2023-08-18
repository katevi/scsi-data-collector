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
	//pollingPeriod = 10 * time.Minutes
	pollingPeriod    = 1 * time.Second
	timeout          = 3 * time.Second
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
		time.Sleep(pollingPeriod)
		timeNow := time.Now().UTC().Format("2006-01-02-15-04-05")
		execDirPath := path.Join(collectedDataDir, timeNow)
		if err := os.Mkdir(execDirPath, os.ModePerm); err != nil {
			appLog.Printf("Failed to create dir %s", err)
		}

		blockDevices, _ := devicelister.GetBlockDevices(allDevicesKey)
		resultsCh := make(chan struct{}, len(blockDevices))
		for _, blockDevice := range blockDevices {
			go execCommand(appLog, resultsCh, execDirPath, blockDevice)
		}

		done := make(chan struct{})
		go func() {
			for i := 0; i < len(blockDevices); i++ {
				<-resultsCh
			}
			done <- struct{}{}
		}()
		select {
		case <-done:
			archiveExecDir(appLog, path.Join(collectedDataDir, timeNow), execDirPath)
		case <-time.After(timeout):
			continue
		}
	}
}

func execCommand(appLog *log.Logger, resultsCh chan<- struct{}, execDirPath string, device devicelister.BlockDevice) {
	cmdArgs := fmt.Sprintf("sg_logs -aa %s", device.Name)
	cmd := exec.Command("bash", "-c", cmdArgs)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		appLog.Printf("Cmd execution completed with err %s %v", err, errb.String())
	}

	deviceName := strings.ReplaceAll(device.Name, "/", "_")
	filepath := path.Join(execDirPath, deviceName)
	dumpCmdResult2File(filepath,
		[]byte(fmt.Sprint("cmd:\n", cmdArgs)),
		[]byte(fmt.Sprint("out:\n", outb.String())),
		[]byte(fmt.Sprint("err:\n", errb.String())))
	resultsCh <- struct{}{}
}

func archiveExecDir(appLog *log.Logger, archiveName, execDir string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("tar czf %s.tar.gz %s --remove-files", archiveName, execDir))
	err := cmd.Run()
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err != nil {
		appLog.Printf("Cmd execution completed with err %s %v", err, errb.String())
	}
}

func dumpCmdResult2File(path string, contents ...[]byte) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, content := range contents {
		f.Write(content)
		f.Write([]byte("\n"))
	}
}
