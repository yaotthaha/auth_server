package main

import (
	"auth_server/lib"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

type ConfigDataStruct struct {
	DB_Conn                  string   `json:"db_conn"`
	DB_Name                  string   `json:"db_name"`
	DB_User                  string   `json:"db_user"`
	DB_Pass                  string   `json:"db_pass"`
	Redis_Conn               string   `json:"redis_conn"`
	Redis_Pass               string   `json:"redis_pass"`
	HTTP_Listen_Address      string   `json:"listen"`
	HTTP_Listen_Port         int      `json:"port"`
	Admin_CIDR_Access_Status bool     `json:"admin_cidr_access_status"`
	Admin_Access_CIDR        []string `json:"admin_access_cidr"`
	Admin_Password           string   `json:"admin_password"`
}

func ConfigFileReadToStruct(configfilename string) (*ConfigDataStruct, int, error) {
	/**
	Request Param:
	@ ConfigFileName string
	Respon Param:
	@ ConfigDataMap map[string]interface{}
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-11) Config File Not Found
	(-12) Config File Read Fail
	(-13) Config File Parse Fail
	*/
	File, err := os.Open(configfilename)
	if err != nil {
		return nil, -11, err
	}
	var FileCloseErr error
	defer func(File *os.File) {
		err = File.Close()
		if err != nil {
			FileCloseErr = err
			return
		}
	}(File)
	if FileCloseErr != nil {
		return nil, -11, FileCloseErr
	}
	ConfigData, err := ioutil.ReadAll(File)
	if err != nil {
		return nil, -11, err
	}
	ConfigParseData := &ConfigDataStruct{}
	err = json.Unmarshal(ConfigData, ConfigParseData)
	if err != nil {
		return nil, -13, err
	}
	return ConfigParseData, 0, nil
}

func ConfigRead(configfile string) (int, error) {
	/**
	Request Param:
	@ ConfigFile string
	Respon Param:
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-14) Config Param `` Not Found
	(-15) Config Param `` Invalid
	*/
	ConfigData, ErrorCode, err := ConfigFileReadToStruct(configfile)
	if ErrorCode != 0 {
		return ErrorCode, err
	}
	TempMap := make(map[string]interface{})
	if ConfigData.DB_Conn != "" {
		DBConnArray := strings.Split(ConfigData.DB_Conn, "://")
		if len(DBConnArray) <= 1 {
			return -15, errors.New("Config Param `db_conn` Invalid")
		} else {
			switch DBConnArray[0] {
			case "tcp":
				DBConnURLArray := strings.Split(DBConnArray[1], ":")
				if !lib.IsNumber(DBConnURLArray[len(DBConnURLArray)-1]) {
					return -15, errors.New("Config Param `db_conn` Invalid")
				}
			case "unix", "file":
				if !lib.IsFileExist(DBConnArray[1]) {
					return -15, errors.New("Config Param `db_conn` Invalid")
				}
			default:
				return -15, errors.New("Config Param `db_conn` Invalid")
			}
		}
		DB_ConnectURL = ConfigData.DB_Conn
		TempMap["db_conn_mode"] = DBConnArray[0]
	}
	if ConfigData.DB_Name != "" {
		if TempMap["db_conn_mode"] != "file" {
			DB_Name = ConfigData.DB_Name
		}
	}
	if ConfigData.DB_User != "" {
		DB_User = ConfigData.DB_User
	}
	if ConfigData.DB_Pass != "" {
		DB_Pass = ConfigData.DB_Pass
	}
	if ConfigData.Redis_Conn != "" {
		RedisConnArray := strings.Split(ConfigData.Redis_Conn, "://")
		if len(RedisConnArray) <= 1 {
			return -15, errors.New("Config Param `redis_conn` Invalid")
		}
		switch RedisConnArray[0] {
		case "tcp":
			RedisConnURLArray := strings.Split(RedisConnArray[1], ":")
			if !lib.IsNumber(RedisConnURLArray[len(RedisConnURLArray)-1]) {
				return -15, errors.New("Config Param `redis_conn` Invalid")
			}
		case "unix":
			if !lib.IsFileExist(RedisConnArray[1]) {
				return -15, errors.New("Config Param `redis_conn` Invalid")
			}
		default:
			return -15, errors.New("Config Param `redis_conn` Invalid")
		}
		Redis_ConnectURL = ConfigData.Redis_Conn
	}
	if ConfigData.Redis_Pass != "" {
		Redis_Pass = ConfigData.Redis_Pass
	}
	if ConfigData.HTTP_Listen_Address != "" {
		HTTP_Listen_Address = ConfigData.HTTP_Listen_Address
	}
	if ConfigData.HTTP_Listen_Port != 0 {
		if lib.IsPortBind("tcp", "", HTTP_Listen_Port, 1) {
			return -15, errors.New("Config Param `listen` or `port` Invalid")
		}
		HTTP_Listen_Port = ConfigData.HTTP_Listen_Port
	}
	Admin_CIDR_Access_Status_Set = ConfigData.Admin_CIDR_Access_Status
	if len(ConfigData.Admin_Access_CIDR) > 0 {
		Admin_Access_CIDR_Set = ConfigData.Admin_Access_CIDR
	} else {
		Admin_Access_CIDR_Set = Admin_Access_CIDR
	}
	if ConfigData.Admin_Password != "" {
		Admin_Password = ConfigData.Admin_Password
	}
	return 0, nil
}
