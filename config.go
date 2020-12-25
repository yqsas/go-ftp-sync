package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type SyncConfig struct {
	IP       string
	Port     int
	Username string
	Password string
	//目标路径
	ScanPath     string
	ScanInterval int
	//并发数
	Concurrency int
	//文件忽略规则
	IgnoreReg string
	//文件上传过程中的标识
	UploadingFlag string
	//delete local file after used
	Delete bool
}

func (jst *SyncConfig) Load(filename string) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, jst)
	if err != nil {
		log.Fatal("load config file failed: ", err)
	}
}

func initSyncConfig() *SyncConfig {
	var syncConfig *SyncConfig
	configFilePath := "./config.json"
	if Exists(configFilePath) {
		syncConfig = new(SyncConfig)
		syncConfig.Load(configFilePath)
		logMsg("load config file success")
	} else {
		logMsg("WARN: config file not found, if you want customized sync config, please create a config file in current directory")
		//初始化配置
		syncConfig = &SyncConfig{
			IP:            "127.0.0.1",
			Port:          21,
			Username:      "admin",
			Password:      "admin",
			ScanPath:      "/home/test",
			ScanInterval:  5,
			Concurrency:   1,
			IgnoreReg:     "[.](inprogress)$",
			UploadingFlag: ".inprogress",
			Delete:        false,
		}
		logMsg("load default config success")
	}
	configContent, _ := json.Marshal(syncConfig)
	logMsg("Config content: ", string(configContent))
	return syncConfig
}
