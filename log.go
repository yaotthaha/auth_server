package main

import (
	"auth_server/lib"
	"database/sql"
	"log"
	"os"
)

var LogFileTag bool = false

func OutputLog(LogLevel int, LogData string) {
	if LogOutputFile != "" && LogFileTag == false {
		_ = os.Remove(LogOutputFile)
		LogFile, err := os.OpenFile(LogOutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if err != nil {
			log.Println("LogFile Read Fail: " + err.Error())
			os.Exit(-1)
		}
		log.SetOutput(LogFile)
		LogFileTag = true
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

func DataBaseLog(DB *sql.DB, Operate, Message string) {
	if !DBLog {
		return
	}
	SQL1 := "INSERT INTO `" + DB_LogTableName + "`("
	SQL2 := ") VALUES ("
	SQL3 := ")"
	SQL1 += "`time`, "
	SQL2 += "'" + "[" + lib.GenDate_String() + "]" + "', "
	//
	SQL1 += "`operate`, "
	SQL2 += "'" + Operate + "', "
	SQL1 += "`message`, "
	SQL2 += "'" + Message + "', "
	//
	SQL1 = SQL1[0 : len(SQL1)-2]
	SQL2 = SQL2[0 : len(SQL2)-2]
	SQL := SQL1 + SQL2 + SQL3
	code, err := lib.DataBaseExec(DB, SQL)
	if code != 0 {
		OutputLog(-2, "DataBase Log Fail: "+err.Error())
	}
}
