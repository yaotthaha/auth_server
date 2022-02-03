package main

import (
	"auth_server/lib"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-redis/redis/v8"
)

func main() {
	SetupCloseHandler()
	ArgsArray := ParamRead()
	if ArgsArray["version"].(bool) {
		fmt.Println(ApplicationName + " " + ApplicationVersion + " (Build From " + ApplicationAuthor + ")")
		return
	}
	ConfigFilePath = ArgsArray["config"].(string)
	DebugMode = ArgsArray["debug_mode"].(bool)
	Admin_CIDR_Access_Status = ArgsArray["access"].(bool)
	LogOutputFile = ArgsArray["logfile"].(string)
	OutputLog(0, ApplicationName+" "+ApplicationVersion+" (Build From "+ApplicationAuthor+")")
	OutputLog(0, "Start...")
	Pre()
	OutputLog(0, "Start HTTP Server...")
	OutputLog(0, "Listen "+HTTP_Listen_Address+":"+strconv.Itoa(HTTP_Listen_Port)+"...")
	HTTPServer(HTTP_Listen_Address, HTTP_Listen_Port)
	return
}

func Pre() {
	OutputLog(0, "Read Config File ("+ConfigFilePath+")...")
	ReadConfigFileStatusCode, ReadConfigFileError := ConfigRead(ConfigFilePath)
	if ReadConfigFileStatusCode != 0 {
		OutputLog(-1, "Read Config File Fail: ["+strconv.Itoa(ReadConfigFileStatusCode)+"] "+ReadConfigFileError.Error())
	}
	OutputLog(0, "Read Config File Success")
	var (
		code int
		err  error
	)
	//
	OutputLog(0, "Connect DataBase...")
	DataBaseConfigMap := make(map[string]string)
	DataBaseConfigMap["url"] = DB_ConnectURL
	DataBaseConfigMap["user"] = DB_User
	DataBaseConfigMap["pass"] = DB_Pass
	DataBaseConfigMap["db_name"] = DB_Name
	DB, code, err = lib.DataBaseConnect(DataBaseConfigMap)
	if code != 0 {
		OutputLog(-1, "Connect DataBase Fail: ["+strconv.Itoa(code)+"] "+err.Error())
	}
	OutputLog(0, "Connect DataBase Success")
	if DBLog {
		SQLCreateLog := "CREATE TABLE IF NOT EXISTS `" + DB_LogTableName + "` (" +
			func() string {
				SQLBasic := ""
				for _, v := range DB_Log_Struct {
					SQLBasic += "`" + v + "` VARCHAR(255) NOT NULL, "
				}
				SQLBasic = SQLBasic[0 : len(SQLBasic)-2]
				return SQLBasic
			}() +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8;"
		code_, _ := lib.DataBaseExec(DB, SQLCreateLog)
		if code_ != 0 {
			OutputLog(-1, "DataBase Prepare Fail")
		}
	}
	SQLCreate := "CREATE TABLE IF NOT EXISTS `" + DB_TableName + "` (" +
		func() string {
			SQLBasic := ""
			for _, v := range DB_Struct {
				SQLBasic += "`" + v + "` VARCHAR(255) NOT NULL, "
			}
			SQLBasic = SQLBasic[0 : len(SQLBasic)-2]
			return SQLBasic
		}() +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	code, err = lib.DataBaseExec(DB, SQLCreate)
	if code != 0 {
		OutputLog(-1, "DataBase Prepare Fail")
	} else {
		OutputLog(0, "DataBase Prepare Success")
	}
	//
	OutputLog(0, "Connect Redis Server...")
	RedisConfigMap := make(map[string]string)
	RedisConfigMap["url"] = Redis_ConnectURL
	RedisConfigMap["pass"] = Redis_Pass
	RedisClient, code, err = lib.RedisConnect(RedisConfigMap)
	if code != 0 {
		OutputLog(-1, "Connect Redis Server Fail: ["+strconv.Itoa(code)+"] "+err.Error())
	}
	OutputLog(0, "Connect Redis Server Success")
}

func ParamRead() map[string]interface{} {
	var (
		insideConfig    string
		insideDebugMode bool
		insideAccessOff bool
		insideLogFile   string
		insideVersion   bool
	)
	flag.StringVar(&insideConfig, "c", ConfigFileDefaultPath, "Config File Path")
	flag.BoolVar(&insideDebugMode, "debug", false, "Debug Mode")
	flag.BoolVar(&insideAccessOff, "accessoff", true, "Access Mode Turn Off")
	flag.StringVar(&insideLogFile, "logfile", "", "Log Output File")
	flag.BoolVar(&insideVersion, "v", false, "Show Version")
	flag.Usage = usage
	flag.Parse()
	argsArray := make(map[string]interface{})
	argsArray["config"] = insideConfig
	argsArray["debug_mode"] = insideDebugMode
	argsArray["access"] = insideAccessOff
	argsArray["logfile"] = insideLogFile
	argsArray["version"] = insideVersion
	return argsArray
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, ApplicationName+` version: `+ApplicationName+`/`+ApplicationVersion+` (Build From `+ApplicationAuthor+`)
Usage: `+ApplicationName+` [-c ConfigFilePath] [-debug]
Options:
   -c {string}  ConfigFilePath (default: `+ConfigFileDefaultPath+`)
   -debug       Debug Mode (default: false)
   -accessoff   Access Mode Turn Off (default: false)
   -logfile     Log Output File (default: none)
`)
}

var interruptTag = false

func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if interruptTag {
			OutputLog(0, "Program has started the end function")
		} else {
			interruptTag = true
			OutputLog(0, "OS Interrupt")
			CloseClean()
			OutputLog(0, "Good Bye!!")
			os.Exit(0)
		}
	}()
}

func CloseClean() {
	_, _ = lib.DataBaseCloseConn(DB)
	func(RedisClient *redis.Client) {
		Keys, _, _ := lib.RedisKeys(RedisClient, Cache_Prefix_Tag+"*")
		for _, v := range Keys {
			_, _ = lib.RedisDel(RedisClient, v)
		}
		return
	}(RedisClient)
	_, _ = lib.RedisConnClose(RedisClient)
}
