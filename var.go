package main

import (
	"database/sql"
	"github.com/go-redis/redis/v8"
	"strings"
)

var (
	DB          *sql.DB
	RedisClient *redis.Client
)

var (
	DB_ConnectURL                string = "tcp://127.0.0.1:3306" // unix:///var/run/mysql.sock // file:///etc/userdata.db
	DB_Name                      string = "auth_user_data"
	DB_User                      string = "1234"
	DB_Pass                      string = "1234"
	Redis_ConnectURL             string = "tcp://127.0.0.1:6379" // unix:///var/run/redis.sock
	Redis_Pass                   string = ""
	HTTP_Listen_Address          string = "0.0.0.0"
	HTTP_Listen_Port             int    = 8082
	Admin_CIDR_Access_Status_Set bool   = true
	Admin_Access_CIDR_Set        []string
	Admin_Password               string = "adminadmin"
)

var (
	DB_TableName                  string = "user_data"
	Cache_Prefix_Tag              string = "auth_"
	Cache_Session_Tag             string = Cache_Prefix_Tag + "session_"
	Cache_Session_Expire_Time     uint64 = 30
	Cache_Token_Expire_Time       uint64 = 7200
	Password_Check_Time           uint64 = 5
	RSA_Bits                      uint64 = 2048
	Cache_Admin_Token_Tag         string = Cache_Prefix_Tag + "admin_token"
	Cache_Admin_Token_Expire_Time uint64 = 600
)

func CacheKeyGen(UserID string, Token string) string {
	str := Cache_Prefix_Tag
	str += "_#_user_id_#_" + UserID
	str += "_#_token_#_" + Token
	return str
}

func CacheKeyToUserIDToken(str string) (string, string) {
	str_ := strings.ReplaceAll(str, Cache_Prefix_Tag, "")
	str_ = strings.ReplaceAll(str_, "_#_user_id_#_", "|")
	str_ = strings.ReplaceAll(str_, "_#_token_#_", "|")
	strArray := strings.Split(str_, "|")
	return strArray[1], strArray[2]
}

var (
	Auth_Open_Status         bool     = true
	Admin_Open_Status        bool     = true
	Admin_CIDR_Access_Status bool     = true
	Admin_Access_CIDR        []string = []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.88.99.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"233.252.0.0/24",
		"240.0.0.0/4",
		"255.255.255.255/32",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}
)

var (
	ApplicationName       string = "AuthServer"
	ApplicationVersion    string = "v2.1.0"
	ApplicationAuthor     string = "Yaott"
	ConfigFileDefaultPath string = "config.json"
	ConfigFilePath        string = ConfigFileDefaultPath
	DebugMode             bool   = false
)

var DB_Struct []string = []string{
	"uuid",
	"username",
	"user_id",
	"status",
	"password_secret",
	"app",
	"expire_time",
	"show_username",
}
