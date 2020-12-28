package main

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var config *SyncConfig
var handlingFiles sync.Map

func main() {
	logMsg("Starting ...")
	config = initSyncConfig()

	var wg sync.WaitGroup
	//start scan goroutine
	fileChan := make(chan string, 10)
	go fileScanner(fileChan)
	//start store goroutine
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go uploader(fileChan)
	}

	wg.Wait()
	logMsg("Bye Bye!")
}

// file scanner
func fileScanner(fileChan chan<- string) {
	logMsg("Start Scanning ...")
	if !Exists(config.ScanPath) {
		if err := os.MkdirAll(config.ScanPath, os.ModeDir); err != nil {
			log.Fatal("cannot create local directory: ", err)
		}
	}
	scanPath := config.ScanPath
	ignoreReg := regexp.MustCompile(config.IgnoreReg)
	for {
		FindFile(fileChan, &scanPath, ignoreReg)
		//sleep
		time.Sleep(time.Second * time.Duration(config.ScanInterval))
	}
}

func uploader(fileChan <-chan string) {
	var ftpClient *ftp.ServerConn
	for filePath := range fileChan {
		for {
			if ftpClient == nil {
				ftpClient = CreateFtpClient(config)
			} else {
				if err := ftpClient.NoOp(); err != nil {
					_ = ftpClient.Quit()
					ftpClient = nil
					continue
				}
				break
			}
			time.Sleep(time.Second * 2)
		}

		if _, loaded := handlingFiles.LoadOrStore(filePath, 0); loaded {
			continue
		}
		if success := Upload(ftpClient, &filePath, config); success && config.Delete {
			_ = os.Remove(filePath)
		}
		handlingFiles.Delete(filePath)
	}
}

//log with goroutine id
func logMsg(msg ...interface{}) {
	log.Println(" ["+GetGoid()+"] ", msg)
}

//get go id
func GetGoid() string {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return strconv.Itoa(id)
}
