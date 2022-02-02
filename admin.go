package main

import (
	"auth_server/lib"
	"encoding/json"
	"strconv"
	"strings"
)

func IsNum(String string) bool {
	_, err := strconv.ParseInt(String, 10, 64)
	if err != nil {
		return false
	} else {
		return true
	}
}

func AcceptString(String string) bool {
	AcceptStr := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890_"
	StrArray := strings.Split(String, "")
	Tag := false
	for _, v := range StrArray {
		if !strings.Contains(AcceptStr, v) {
			Tag = true
		}
	}
	if Tag {
		return false
	} else {
		return true
	}
}

func AdminRoute(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ ParamMap map[string]string
	@ ConnectionTag string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(111) Param `` Invalid
	(...)
	*/
	if ParamMap["method"] == "" {
		OutputLog(0, "[HTTP Server] [Admin] ["+ConnectionTag+"] Method Not Found")
		return HTTPJSONRespon("110", "Param `method` Not Found")
	}
	switch ParamMap["method"] {
	case "login":
		return AdminLogin(ParamMap, ConnectionTag)
	case "logout":
		return AdminLogout(ParamMap, ConnectionTag)
	case "user_add":
		delete(ParamMap, "method")
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminUserAdd(ParamMap, ConnectionTag)
	case "user_del":
		delete(ParamMap, "method")
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminUserDel(ParamMap, ConnectionTag)
	case "user_set":
		delete(ParamMap, "method")
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminUserSet(ParamMap, ConnectionTag)
	case "user_list":
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminUserList(ConnectionTag)
	case "auth_close":
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminCloseAuth(ConnectionTag)
	case "admin_close":
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminCloseAdmin(ConnectionTag)
	case "auth_open":
		if ResponHTTP, CheckAccess := AdminAccessCheck(ParamMap, ConnectionTag); !CheckAccess {
			return ResponHTTP
		} else {
			delete(ParamMap, "admin_token")
		}
		return AdminOpenAuth(ConnectionTag)
	default:
		OutputLog(0, "[HTTP Server] [Admin] ["+ConnectionTag+"] Method Not Found")
		return HTTPJSONRespon("111", "Param `method` Invalid")
	}
}

func AdminLogin(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ admin_password string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(113) Admin Password Invalid
	(198) API Call Fail
	*/
	if ParamMap["admin_password"] == "" {
		OutputLog(0, lib.JoinString("[Admin Login] [", ConnectionTag, "] Param `admin_password` Not Found"))
		return HTTPJSONRespon("110", "Param `admin_password` Not Found")
	}
	LoginTag := false
	TimeStampNowString := lib.GetTimestamp_S_String()
	for i := 0; i < 5; i++ {
		TimeStampTemp := lib.BigIntReduce(TimeStampNowString, strconv.Itoa(i))
		SecretInfoArray := strings.Split(ParamMap["admin_password"], "_TIMESTAMP_")
		if SecretInfoArray[1] != TimeStampTemp {
			continue
		}
		if SecretInfoArray[0] == Admin_Password {
			LoginTag = true
			break
		}
	}
	if !LoginTag {
		OutputLog(0, lib.JoinString("[Admin Login] [", ConnectionTag, "] Admin Password Invalid"))
		return HTTPJSONRespon("113", "Admin Password Invalid")
	}
	CheckAdminToken, code, err := lib.RedisGet(RedisClient, Cache_Admin_Token_Tag)
	if code != 0 && code != -26 {
		OutputLog(-2, lib.JoinString("[Admin Login] [", ConnectionTag, "] Error Redis Get Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	var Token string
	if code == -26 {
		Token = lib.GenRandomString(32)
		code, err = lib.RedisSet(RedisClient, Cache_Admin_Token_Tag, Token, int64(Cache_Admin_Token_Expire_Time))
		if code != 0 {
			OutputLog(-2, lib.JoinString("[Admin Login] [", ConnectionTag, "] Error Redis Set Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		}
	} else {
		Token = CheckAdminToken
		code, err = lib.RedisExpire(RedisClient, Cache_Admin_Token_Tag, int64(Cache_Admin_Token_Expire_Time))
		if code != 0 {
			OutputLog(-2, lib.JoinString("[Admin Login] [", ConnectionTag, "] Error Redis Set Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		}
	}
	ReturnMap := make(map[string]interface{})
	ReturnMap["admin_token"] = Token
	return HTTPJSONRespon("100", ReturnMap)
}

func AdminLogout(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ admin_token string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(114) Admin Token Auth Fail
	(115) Admin Token Not Found
	(198) API Call Fail
	*/
	if ParamMap["admin_token"] == "" {
		OutputLog(0, lib.JoinString("[Admin Logout] [", ConnectionTag, "] Param `admin_token` Not Found"))
		return HTTPJSONRespon("110", "Param `admin_token` Not Found")
	}
	CheckAdminToken, code, err := lib.RedisGet(RedisClient, Cache_Admin_Token_Tag)
	if code != 0 && code != -26 {
		OutputLog(-2, lib.JoinString("[Admin Logout] [", ConnectionTag, "] Error Redis Get Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	if code == -26 {
		OutputLog(0, lib.JoinString("[Admin Logout] [", ConnectionTag, "] Admin Token Not Found"))
		return HTTPJSONRespon("115", "Admin Token Not Found")
	}
	if CheckAdminToken == ParamMap["admin_token"] {
		code, err = lib.RedisDel(RedisClient, Cache_Admin_Token_Tag)
		if code != 0 {
			OutputLog(-2, lib.JoinString("[Admin Logout] [", ConnectionTag, "] Error Redis Del Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		} else {
			OutputLog(0, lib.JoinString("[Admin Logout] [", ConnectionTag, "] Admin Logout"))
			return HTTPJSONRespon("100", "OK")
		}
	} else {
		OutputLog(0, lib.JoinString("[Admin Logout] [", ConnectionTag, "] Admin Token Auth Fail"))
		return HTTPJSONRespon("114", "Admin Token Auth Fail")
	}
}

func AdminAccessCheck(ParamMap map[string]string, ConnectionTag string) (string, bool) {
	/**
	Request Param:
	@ admin_token string
	(...)
	Respon Param:
	@ HTTPRespon string
	@ Access bool
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(114) Admin Token Auth Fail
	(198) API Call Fail
	*/
	if ParamMap["admin_token"] == "" {
		OutputLog(0, lib.JoinString("[Admin Access] [", ConnectionTag, "] Param `admin_token` Not Found"))
		return HTTPJSONRespon("110", "Param `admin_token` Not Found"), false
	}
	TokenCache, code, err := lib.RedisGet(RedisClient, Cache_Admin_Token_Tag)
	if code != 0 {
		if code == -26 {
			OutputLog(0, lib.JoinString("[Admin Access] [", ConnectionTag, "] Admin Token Auth Fail"))
			return HTTPJSONRespon("114", "Admin Token Auth Fail"), false
		} else {
			OutputLog(0, lib.JoinString("[Admin Access] [", ConnectionTag, "] Error Redis Get Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail"), false
		}
	}
	if TokenCache == ParamMap["admin_token"] {
		return "nil", true
	} else {
		OutputLog(0, lib.JoinString("[Admin Access] [", ConnectionTag, "] Admin Token Auth Fail"))
		return HTTPJSONRespon("114", "Admin Token Auth Fail"), false
	}
}

func AdminUserAdd(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ username string
	@ user_id string
	@ status string(bool)
	@ password string
	@ app string
	@ expire_time string(uint64)
	@ show_username string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(111) Param `` Invalid
	(198) API Call Fail
	*/
	WriteMap := make(map[string]string)
	// username
	if ParamMap["username"] == "" {
		OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `username` Not Found"))
		return HTTPJSONRespon("110", "Param `username` Not Found")
	} else {
		if !AcceptString(ParamMap["username"]) {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `username` Invalid: `username` contains illegal characters"))
			return HTTPJSONRespon("111", "Param `username` Invalid: `username` contains illegal characters")
		}
		CheckUsernameExist, code, err := DataBaseCheckFieldExist(DB, "username", ParamMap["username"])
		if code != 0 {
			OutputLog(-2, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Error Database Check `username` Exist Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		}
		if CheckUsernameExist {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `username` Invalid: `username` already exists"))
			return HTTPJSONRespon("111", "Param `username` Invalid: `username` already exists")
		}
		WriteMap["username"] = ParamMap["username"]
	}
	// user_id
	if ParamMap["user_id"] == "" {
		OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `user_id` Not Found"))
		return HTTPJSONRespon("110", "Param `user_id` Not Found")
	} else {
		if !AcceptString(ParamMap["user_id"]) {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `user_id` Invalid: `user_id` contains illegal characters"))
			return HTTPJSONRespon("111", "Param `user_id` Invalid: `user_id` contains illegal characters")
		}
		CheckUseridExist, code, err := DataBaseCheckFieldExist(DB, "user_id", ParamMap["user_id"])
		if code != 0 {
			OutputLog(-2, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Error Database Check `user_id` Exist Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		}
		if CheckUseridExist {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `user_id` Invalid: `user_id` already exists"))
			return HTTPJSONRespon("111", "Param `user_id` Invalid: `user_id` already exists")
		}
		WriteMap["user_id"] = ParamMap["user_id"]
	}
	// status
	if ParamMap["status"] != "" {
		if !(ParamMap["status"] == "true" || ParamMap["status"] == "false") {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `status` Invalid: `status` only support `true` or `false` (string)"))
			return HTTPJSONRespon("111", "Param `status` Invalid: `status` only support `true` or `false` (string)")
		}
		WriteMap["status"] = ParamMap["status"]
	} else {
		WriteMap["status"] = "true"
	}
	// password
	if ParamMap["password"] != "" {
		if !AcceptString(ParamMap["password"]) {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `password` Invalid: `password` contains illegal characters"))
			return HTTPJSONRespon("111", "Param `password` Invalid: `password` contains illegal characters")
		}
		WriteMap["password_secret"] = lib.Sha256Sum(ParamMap["password"])
	} else {
		OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `password` Not Found"))
		return HTTPJSONRespon("110", "Param `password` Not Found")
	}
	// app
	if ParamMap["app"] == "" {
		OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `app` Not Found"))
		return HTTPJSONRespon("110", "Param `app` Not Found")
	} else {
		if strings.Contains(ParamMap["app"], ",") {
			AppArray := strings.Split(ParamMap["app"], ",")
			for _, v := range AppArray {
				if !AcceptString(v) {
					OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `app` Invalid: `app` contains illegal characters"))
					return HTTPJSONRespon("111", "Param `app` Invalid: `app` contains illegal characters")
				}
			}
			AppArrayJSON, _ := json.Marshal(AppArray)
			WriteMap["app"] = lib.StringToBase64(string(AppArrayJSON))
		} else {
			if !AcceptString(ParamMap["app"]) {
				OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `app` Invalid: `app` contains illegal characters"))
				return HTTPJSONRespon("111", "Param `app` Invalid: `app` contains illegal characters")
			}
			WriteMap["app"] = lib.StringToBase64(lib.JoinString(`["`, ParamMap["app"], `"]`))
		}
	}
	// expire_time
	if ParamMap["expire_time"] != "" {
		if !IsNum(ParamMap["expire_time"]) {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `expire_time` Invalid: `expire_time` not a valid timestamp"))
			return HTTPJSONRespon("111", "Param `expire_time` Invalid: `expire_time` not a valid timestamp")
		}
		if lib.StringNumCompare(ParamMap["expire_time"], lib.GetTimestamp_S_String()) < 0 {
			OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `expire_time` Invalid: `expire_time` not a valid timestamp"))
			return HTTPJSONRespon("111", "Param `expire_time` Invalid: `expire_time` not a valid timestamp")
		}
		WriteMap["expire_time"] = ParamMap["expire_time"]
	} else {
		WriteMap["expire_time"] = "0"
	}
	// show_username
	if ParamMap["show_username"] == "" {
		WriteMap["show_username"] = ParamMap["username"]
	} else {
		WriteMap["show_username"] = ParamMap["show_username"]
	}
	CheckShowUsernameExist, code, err := DataBaseCheckFieldExist(DB, "show_username", WriteMap["show_username"])
	if code != 0 {
		OutputLog(-2, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Error Database Check `show_username` Exist Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	if CheckShowUsernameExist {
		OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `show_username` Invalid: `show_username` already exists"))
		return HTTPJSONRespon("111", "Param `show_username` Invalid: `show_username` already exists")
	}
	code, err = DataBaseUserAdd(DB, WriteMap["user_id"], WriteMap["username"], WriteMap["status"], WriteMap["password_secret"], WriteMap["app"], WriteMap["expire_time"], WriteMap["show_username"])
	if code != 0 {
		OutputLog(-2, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Error DataBase Exec Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	return HTTPJSONRespon("100", "OK")
}

func AdminUserDel(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ user_id string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(111) Param `` Invalid
	(112) User Not Found
	(198) API Call Fail
	*/
	var UserInfo map[string]string
	if ParamMap["user_id"] == "" {
		OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Param `user_id` Not Found"))
		return HTTPJSONRespon("110", "Param `user_id` Not Found")
	} else {
		GetUserInfo, code, err := DataBaseSearchUserInfo(DB, ParamMap["user_id"])
		if code != 0 {
			if code == -51 {
				OutputLog(0, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] User Not Found"))
				return HTTPJSONRespon("112", "User Not Found")
			}
			OutputLog(-2, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Error Database Check `user_id` Exist Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		}
		UserInfo = GetUserInfo
	}
	code, err := DataBaseUserDel(DB, UserInfo["uuid"])
	if code != 0 {
		OutputLog(-2, lib.JoinString("[Admin UserAdd] [", ConnectionTag, "] Error DataBase Exec Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	return HTTPJSONRespon("100", "OK")
}

func AdminUserSet(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ user_id_index string
	@ *user_id string
	@ *username string
	@ *status string
	@ *app string
	@ *password string
	@ *expire_time string
	@ *show_username string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(100) OK
	(110) Param `` Not Found
	(111) Param `` Invalid
	(112) User Not Found
	(198) API Call Fail
	*/
	var UserIndex string
	if ParamMap["user_id_index"] == "" {
		OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `user_id_index` Not Found"))
		return HTTPJSONRespon("110", "Param `user_id_index` Not Found")
	} else {
		GetUserInfo, code, err := DataBaseSearchUserInfo(DB, ParamMap["user_id_index"])
		if code != 0 {
			if code == -51 {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] User Not Found"))
				return HTTPJSONRespon("112", "User Not Found")
			}
			OutputLog(-2, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Error Database Check `user_id_index` Exist Fail: ", err.Error()))
			return HTTPJSONRespon("198", "API Call Fail")
		}
		UserIndex = GetUserInfo["uuid"]
	}
	delete(ParamMap, "user_id_index")
	InputMap := make(map[string]string)
	for k, v := range ParamMap {
		switch k {
		case "user_id":
			if !AcceptString(v) {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `user_id` Invalid: `user_id` contains illegal characters"))
				return HTTPJSONRespon("111", "Param `user_id` Invalid: `user_id` contains illegal characters")
			}
			CheckExist, code, err := DataBaseCheckFieldExist(DB, "user_id", v)
			if code != 0 {
				OutputLog(-2, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Error DataBase Query Fail: "+err.Error()))
				return HTTPJSONRespon("198", "API Call Fail")
			}
			if CheckExist {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `user_id` Invalid: `user_id` already exists: ", err.Error()))
				return HTTPJSONRespon("110", "Param `user_id` Invalid: `user_id` already exists")
			}
			InputMap[k] = v
		case "username":
			if !AcceptString(v) {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `username` Invalid: `username` contains illegal characters"))
				return HTTPJSONRespon("111", "Param `username` Invalid: `username` contains illegal characters")
			}
			CheckExist, code, err := DataBaseCheckFieldExist(DB, "username", v)
			if code != 0 {
				OutputLog(-2, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Error DataBase Query Fail: "+err.Error()))
				return HTTPJSONRespon("198", "API Call Fail")
			}
			if CheckExist {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `username` Invalid: `username` already exists: ", err.Error()))
				return HTTPJSONRespon("110", "Param `username` Invalid: `username` already exists")
			}
			InputMap[k] = v
		case "status":
			if !(v == "true" || v == "false") {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `status` Invalid: only support `true` or `false`"))
				return HTTPJSONRespon("111", "Param `status` Invalid: only support `true` or `false`")
			}
			InputMap[k] = v
		case "app":
			if strings.Contains(v, ",") {
				AppArray := strings.Split(v, ",")
				for _, v2 := range AppArray {
					if !AcceptString(v2) {
						OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `app` Invalid: `app` contains illegal characters"))
						return HTTPJSONRespon("111", "Param `app` Invalid: `app` contains illegal characters")
					}
				}
				AppArrayJSON, _ := json.Marshal(AppArray)
				InputMap[k] = lib.StringToBase64(string(AppArrayJSON))
			} else {
				if !AcceptString(ParamMap["app"]) {
					OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `app` Invalid: `app` contains illegal characters"))
					return HTTPJSONRespon("111", "Param `app` Invalid: `app` contains illegal characters")
				}
				InputMap[k] = lib.StringToBase64(lib.JoinString(`["`, v, `"]`))
			}
		case "password":
			if !AcceptString(v) {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `password` Invalid: `password` contains illegal characters"))
				return HTTPJSONRespon("111", "Param `password` Invalid: `password` contains illegal characters")
			}
			InputMap["password_secret"] = lib.Sha256Sum(v)
		case "expire_time":
			if !IsNum(v) {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `expire_time` Invalid: `expire_time` not a valid timestamp"))
				return HTTPJSONRespon("111", "Param `expire_time` Invalid: `expire_time` not a valid timestamp")
			}
			if lib.StringNumCompare(v, lib.GetTimestamp_S_String()) < 0 {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `expire_time` Invalid: `expire_time` not a valid timestamp"))
				return HTTPJSONRespon("111", "Param `expire_time` Invalid: `expire_time` not a valid timestamp")
			}
			InputMap["expire_time"] = v
		case "show_username":
			if !AcceptString(v) {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `show_username` Invalid: `show_username` contains illegal characters"))
				return HTTPJSONRespon("111", "Param `show_username` Invalid: `show_username` contains illegal characters")
			}
			CheckExist, code, err := DataBaseCheckFieldExist(DB, "show_username", v)
			if code != 0 {
				OutputLog(-2, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Error DataBase Query Fail: "+err.Error()))
				return HTTPJSONRespon("198", "API Call Fail")
			}
			if CheckExist {
				OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param `show_username` Invalid: `show_username` already exists: ", err.Error()))
				return HTTPJSONRespon("110", "Param `show_username` Invalid: `show_username` already exists")
			}
			InputMap[k] = v
		default:
		}
	}
	if len(InputMap) <= 0 {
		OutputLog(0, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Param Num Invalid"))
		return HTTPJSONRespon("111", "Param Num Invalid")
	}
	code, err := DataBaseUserSet(DB, UserIndex, InputMap)
	if code != 0 {
		OutputLog(-2, lib.JoinString("[Admin UserSet] [", ConnectionTag, "] Error DataBase Exec Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	return HTTPJSONRespon("100", "OK")
}

func AdminUserList(ConnectionTag string) string {
	Result, code, err := DataBaseUserList(DB)
	if code != 0 {
		OutputLog(-2, lib.JoinString("[Admin UserList] [", ConnectionTag, "] Error Database Query Fail: ", err.Error()))
		return HTTPJSONRespon("198", "API Call Fail")
	}
	return HTTPJSONRespon("100", Result)
}

func AdminCloseAdmin(ConnectionTag string) string {
	/**
	Request Param:
	Respon Param:
	@ StatusCode int
	Statuscode:
	(100) OK
	*/
	Admin_Open_Status = false
	OutputLog(1, lib.JoinString("[Admin Option] [", ConnectionTag, "] Admin Close Admin API"))
	return HTTPJSONRespon("100", "OK")
}

func AdminCloseAuth(ConnectionTag string) string {
	/**
	Request Param:
	Respon Param:
	@ StatusCode int
	Statuscode:
	(100) OK
	*/
	Auth_Open_Status = false
	OutputLog(1, lib.JoinString("[Admin Option] [", ConnectionTag, "] Admin Close Auth API"))
	return HTTPJSONRespon("100", "OK")
}

func AdminOpenAuth(ConnectionTag string) string {
	/**
	Request Param:
	Respon Param:
	@ StatusCode int
	Statuscode:
	(100) OK
	*/
	Auth_Open_Status = true
	OutputLog(1, lib.JoinString("[Admin Option] [", ConnectionTag, "] Admin Reopen Auth API"))
	return HTTPJSONRespon("100", "OK")
}
