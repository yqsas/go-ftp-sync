package main

import (
	_ "errors"
	"fmt"
	"github.com/jlaffaye/ftp"
	"io/ioutil"
	_ "io/ioutil"
	"log"
	"os"
	"path"
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
	//start scan oroutine
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
		findFile(fileChan, &scanPath, ignoreReg)
		//扫描间隔等待
		time.Sleep(time.Second * time.Duration(config.ScanInterval))
	}
}

//find all file
func findFile(fileChan chan<- string, dir *string, ignoreRegx *regexp.Regexp) {
	files, _ := ioutil.ReadDir(*dir)
	for _, f := range files {
		filePath := *dir + "/" + f.Name()
		if f.IsDir() {
			findFile(fileChan, &filePath, ignoreRegx)
			continue
		}
		if match := ignoreRegx.MatchString(filePath); match {
			logMsg("ignore file: " + filePath)
			continue
		}
		fileChan <- filePath
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
		if success := upload(ftpClient, &filePath, config); success && config.Delete {
			_ = os.Remove(filePath)
		}
		handlingFiles.Delete(filePath)
	}
}

func upload(client *ftp.ServerConn, filePath *string, config *SyncConfig) bool {
	logMsg("file upload start：" + *filePath)

	if !Exists(*filePath) {
		logMsg("file is not exists: ", *filePath)
		return false
	}
	toSend, err := os.Open(*filePath)
	if toSend != nil {
		defer toSend.Close()
	}
	if err != nil {
		logMsg("file open filed: ", err)
		return false
	}
	filename := path.Base(*filePath)
	parentDir := path.Dir(strings.TrimPrefix(*filePath, config.ScanPath))
	if err := client.ChangeDir(parentDir); err != nil {
		if err := Mkdir(client, parentDir); err != nil {
			return false
		}
	}
	//append file to ftp with uploadingFlag
	err = client.Append(filename+config.UploadingFlag, toSend)

	if err != nil {
		logMsg("store file error: ", err)
		return false
	}
	//rename file to origin name
	if err = client.Rename(filename+config.UploadingFlag, filename); err != nil {
		logMsg("rename file error:", err)
		return false
	}
	logMsg("file upload success：" + *filePath)
	return true
}

func logMsg(msg ...interface{}) {
	log.Println(" ["+GetGoid()+"] ", msg)
}

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
