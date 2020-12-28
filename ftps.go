package main

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"os"
	"path"
	"strings"
	"time"
)

// create a new ftp client
func CreateFtpClient(config *SyncConfig) (ftpClient *ftp.ServerConn) {
	var addr = fmt.Sprintf("%s:%d", strings.TrimPrefix(config.IP, "ftp://"), config.Port)

	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		logMsg("FTP connect failed", err)
		return nil
	}

	err = c.Login(config.Username, config.Password)
	if err != nil {
		logMsg("FTP login failed", err)
		_ = c.Quit()
		return nil
	}

	return c
}

// make ftp directory
func Mkdir(client *ftp.ServerConn, dir string) error {
	var newPath string
	var nestPath string

	if i := strings.Index(dir, "/"); i != -1 {
		newPath = dir[:i+1]
		nestPath = dir[i+1:]
	} else {
		newPath = dir
	}

	if err := client.ChangeDir(newPath); err != nil {
		if err := client.MakeDir(newPath); err != nil {
			time.Sleep(time.Second * 2) //防止并发错误，再尝试一次
			if err := client.ChangeDir(newPath); err != nil {
				logMsg("FTP directory create failed: ", err)
				return err
			}
		}
		_ = client.ChangeDir(newPath)
	}
	if nestPath != "" {
		return Mkdir(client, nestPath)
	} else {
		logMsg("FTP directory create sucess: " + newPath)
		return nil
	}
}

func Upload(client *ftp.ServerConn, filePath *string, config *SyncConfig) bool {
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
