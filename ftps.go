package main

import (
	"fmt"
	"github.com/jlaffaye/ftp"
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
