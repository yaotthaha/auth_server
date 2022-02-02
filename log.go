package main

import (
	"auth_server/lib"
	"log"
	"os"
)

func OutputLog(LogLevel int, LogData string) {
	if LogOutputFile != "" {
		LogFile, err := os.OpenFile(LogOutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if err != nil {
			log.Println("LogFile Read Fail: " + err.Error())
			os.Exit(-1)
		}
		log.SetOutput(LogFile)
	}
	log.SetFlags(0)
	var LogLevelTag string
	switch LogLevel {
	case 0:
		LogLevelTag = "Info"
	case 1:
		LogLevelTag = "Warning"
	case 2:
		if !DebugMode {
			return
		}
		LogLevelTag = "Debug"
	case -1:
		LogLevelTag = "Fatal Error"
	case -2:
		LogLevelTag = "Error"
	default:
		LogLevelTag = "Info"
	}
	log.SetPrefix("[" + lib.GenDate_String() + "] [" + LogLevelTag + "] ")
	if LogLevel == -1 {
		log.Println(LogData)
		CloseClean()
		os.Exit(-1)
	} else {
		log.Println(LogData)
	}
}
