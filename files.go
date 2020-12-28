package main

import (
	"io/ioutil"
	"os"
	"regexp"
)

//find all file
func FindFile(fileChan chan<- string, dir *string, ignoreRegx *regexp.Regexp) {
	files, _ := ioutil.ReadDir(*dir)
	for _, f := range files {
		filePath := *dir + "/" + f.Name()
		if f.IsDir() {
			FindFile(fileChan, &filePath, ignoreRegx)
			continue
		}
		if match := ignoreRegx.MatchString(filePath); match {
			logMsg("ignore file: " + filePath)
			continue
		}
		fileChan <- filePath
	}
}

// path is exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
