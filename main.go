package main

import (
	"29go_redis/config"
	"29go_redis/lib/logger"
	"29go_redis/resp/handler"
	"29go_redis/tcp"
	"fmt"
	"os"
)

// ACE : http://patorjk.com/software/taag/#p=testall&h=2&v=2&f=Graffiti&t=ACE
var banner = `
======================================================
     _       ____   _____           ____    _____   ____    ___   ____  
    / \     / ___| | ____|         |  _ \  | ____| |  _ \  |_ _| / ___| 
   / _ \   | |     |  _|    _____  | |_) | |  _|   | | | |  | |  \___ \ 
  / ___ \  | |___  | |___  |_____| |  _ <  | |___  | |_| |  | |   ___) |
 /_/   \_\  \____| |_____|         |_| \_\ |_____| |____/  |___| |____/
======================================================
`

const configFile string = "redis.conf"

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	print(banner)
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{Address: fmt.Sprintf("%s:%d", config.Properties.Bind,
			config.Properties.Port)},
		handler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
