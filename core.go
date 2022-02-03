package main

import (
	"auth_server/lib"
	"encoding/json"
	"strconv"
	"strings"
)

func HTTPJSONRespon(code string, msg interface{}) string {
	HTTPResponMap := make(map[string]interface{})
	HTTPResponMap["code"] = code
	HTTPResponMap["msg"] = msg
	HTTPResponJSON, _ := json.Marshal(HTTPResponMap)
	return string(HTTPResponJSON)
}

func IsAccess(AppArray []string, AppName string) bool {
	if len(AppArray) <= 0 {
		return false
	}
	if AppArray[0] == "all" {
		return true
	}
	for _, v := range AppArray {
		if v == AppName {
			return true
		}
	}
	return false
}

func AuthRoute(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ ParamMap map[string]string
	@ ConnectionTag string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(000) OK
	(010) Param `` Not Found
	(011) Param `` Invalid
	ParamMap:
	"method" ==> "user_pre/user_auth/app_auth/token_exist"
	"..." ==> "..."
	*/
	if ParamMap["method"] == "" {
		OutputLog(0, "[HTTP Server] [Auth] ["+ConnectionTag+"] Method Not Found")
		return HTTPJSONRespon("010", "Param `method` Not Found")
	}
	switch ParamMap["method"] {
	case "user_prepare":
		return AuthUserPrepare(ConnectionTag)
	case "user_auth":
		delete(ParamMap, "method")
		return AuthUserAuth(ParamMap, ConnectionTag)
	case "app_auth":
		delete(ParamMap, "method")
		return AuthAppAuth(ParamMap, ConnectionTag)
	case "token_exist":
		delete(ParamMap, "method")
		return AuthTokenExist(ParamMap, ConnectionTag)
	default:
		OutputLog(0, "[HTTP Server] [Auth] ["+ConnectionTag+"] Method Not Found")
		return HTTPJSONRespon("011", "Param `method` Invalid")
	}
}

func AuthUserPrepare(ConnectionTag string) string {
	/**
	Request Param:
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(000) OK
	(010) Param `` Not Found
	(011) Param `` Invalid
	(098) API Call Fail
	(099) Others Error
	ReturnMap:
	@ session_id string
	@ timestamp string
	@ public_key string
	*/
	NewSessionID := lib.GenRandomString(32)
	OutputLog(0, lib.JoinString("[User Prepare] [", ConnectionTag, "] New Session ID: ", NewSessionID))
	PriKey, PubKey, err := lib.RSAGen(RSA_Bits)
	if err != nil {
		OutputLog(2, "[User Prepare] Error: RSA Generate Fail: "+err.Error())
		return HTTPJSONRespon("099", "RSA Generate Fail")
	}
	OutputLog(0, lib.JoinString("[User Prepare] [", ConnectionTag, "] Generate RSA"))
	SaveRedisCode, SaveRedisErr := lib.RedisSet(RedisClient, Cache_Session_Tag+NewSessionID, string(PriKey), int64(Cache_Session_Expire_Time))
	if SaveRedisCode != 0 {
		OutputLog(-2, "[User Prepare] Redis Save Fail: "+SaveRedisErr.Error())
		return HTTPJSONRespon("098", "API Call Fail")
	}
	OutputLog(2, lib.JoinString("[User Prepare] [", ConnectionTag, "] Save RSA PriKey To Redis Server"))
	ReturnMap := make(map[string]string)
	ReturnMap["session_id"] = NewSessionID
	ReturnMap["timestamp"] = lib.GetTimestamp_S_String()
	ReturnMap["public_key"] = string(PubKey)
	return HTTPJSONRespon("000", ReturnMap)
}

func AuthUserAuth(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ ParamMap map[string]string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(000) OK
	(010) Param `` Not Found
	(011) Param `` Invalid
	(012) Session Invalid
	(013) Password Invalid
	(014) User Not Found
	(015) User Expired
	(016) User Disable
	(017) Auth Timeout
	(019) User Access Denied
	(098) API Call Fail
	(099) Others Error
	ParamMap:
	@ session_id
	@ secret
	@ app_name
	ReturnMap:
	@ token string
	@* app_access bool
	*/
	if ParamMap["session_id"] == "" {
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] Session ID Not Found"))
		return HTTPJSONRespon("010", "Param `session_id` Not Found")
	}
	if ParamMap["secret"] == "" {
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] Param `secret` Not Found"))
		return HTTPJSONRespon("010", "Param `secret` Not Found")
	}
	GetPriKey, code, err := lib.RedisGet(RedisClient, Cache_Session_Tag+ParamMap["session_id"])
	if code != 0 {
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] Session ID Not Found"))
		return HTTPJSONRespon("012", "Session Invalid")
	}
	DecryptInfo := make(map[string]string)
	var UserInfoRespon map[string]string
	RealSecretInfoByte, err := lib.RSADecrypt([]byte(lib.Base64ToString(ParamMap["secret"])), []byte(GetPriKey))
	if err != nil {
		OutputLog(-2, lib.JoinString("[User Auth] [", ConnectionTag, "] RSA Decrypt Fail: ", err.Error()))
		return HTTPJSONRespon("013", "Password Invalid")
	}
	TimeStampNowString := lib.GetTimestamp_S_String()
	for i := 0; i < int(Password_Check_Time); i++ {
		TimeStampTemp := lib.BigIntReduce(TimeStampNowString, strconv.Itoa(i))
		SecretInfoArray := strings.Split(string(RealSecretInfoByte), "_TIMESTAMP_")
		if SecretInfoArray[2] != TimeStampTemp {
			continue
		}
		DecryptInfo["UserIDInput"] = SecretInfoArray[1]
		DecryptInfo["PasswordSecretInput"] = SecretInfoArray[3]
		break
	}
	if len(DecryptInfo) <= 0 {
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] SecretInfo Not Found"))
		return HTTPJSONRespon("017", "Auth Timeout")
	}
	UserInfoRespon, code, err = DataBaseSearchUserInfo(DB, DecryptInfo["UserIDInput"])
	if code != 0 {
		DataBaseLog(DB, "用户登录", "状态：失败 | 额外信息：登入的用户ID{"+DecryptInfo["UserIDInput"]+"}未找到")
		OutputLog(-2, lib.JoinString("[User Auth] [", ConnectionTag, "] DataBase Search Fail: ", err.Error()))
		return HTTPJSONRespon("013", "Password Invalid")
	}
	if len(UserInfoRespon) <= 0 {
		DataBaseLog(DB, "用户登录", "状态：失败 | 额外信息：用户ID{"+DecryptInfo["UserIDInput"]+"}未找到")
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] User Not Found"))
		return HTTPJSONRespon("014", "User Not Found")
	}
	if UserInfoRespon["status"] != "true" {
		DataBaseLog(DB, "用户登录", "状态：失败 | 额外信息：用户ID{"+UserInfoRespon["user_id"]+"}已禁用")
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] User Disable"))
		return HTTPJSONRespon("016", "User Disable")
	}
	if UserInfoRespon["expire_time"] != "0" {
		if lib.StringNumCompare(UserInfoRespon["expire_time"], lib.GetTimestamp_S_String()) < 0 {
			DataBaseLog(DB, "用户登录", "状态：失败 | 额外信息：用户ID{"+UserInfoRespon["user_id"]+"}已过期 | 过期时间：["+lib.TimestampToDate(func(TimestampString string) int64 {
				t, _ := strconv.ParseInt(TimestampString, 10, 64)
				return t
			}(UserInfoRespon["expire_time"]))+"]")
			OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] User Expired"))
			return HTTPJSONRespon("015", "User Expired")
		}
	}
	if DecryptInfo["PasswordSecretInput"] != UserInfoRespon["password_secret"] {
		DataBaseLog(DB, "用户登录", "状态：失败 | 额外信息：用户ID{"+UserInfoRespon["user_id"]+"} 密码错误")
		OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] Password Invalid"))
		return HTTPJSONRespon("013", "Password Invalid")
	}
	var Token string
	GetRedisKeyArray, code, err := lib.RedisKeys(RedisClient, CacheKeyGen(DecryptInfo["UserIDInput"], "*"))
	if code != 0 {
		Token = lib.GenRandomString(32)
		UserInfoResponString, _ := json.Marshal(UserInfoRespon)
		RedisSetCode, RedisSetError := lib.RedisSet(RedisClient, CacheKeyGen(DecryptInfo["UserIDInput"], Token), string(UserInfoResponString), int64(Cache_Token_Expire_Time))
		if RedisSetCode != 0 {
			OutputLog(-2, lib.JoinString("[User Auth] [", ConnectionTag, "] Redis Set Fail: "+RedisSetError.Error()))
			return HTTPJSONRespon("098", "API Call Fail")
		}
	} else {
		_, Token = CacheKeyToUserIDToken(GetRedisKeyArray[0])
		RedisExpireCode, RedisExpireError := lib.RedisExpire(RedisClient, GetRedisKeyArray[0], int64(Cache_Token_Expire_Time))
		if RedisExpireCode != 0 {
			OutputLog(-2, lib.JoinString("[User Auth] [", ConnectionTag, "] Redis Expire Fail: "+RedisExpireError.Error()))
			return HTTPJSONRespon("098", "API Call Fail")
		}
	}
	ReturnMap := make(map[string]interface{})
	if ParamMap["app_name"] != "" {
		var AppArray []string
		_ = json.Unmarshal([]byte(lib.Base64ToString(UserInfoRespon["app"])), &AppArray)
		if IsAccess(AppArray, ParamMap["app_name"]) {
			ReturnMap["app_access"] = true
		} else {
			OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] User Access Denied, AppName: {", ParamMap["app_name"], "}"))
			return HTTPJSONRespon("019", "User Access Denied")
		}
	}
	DataBaseLog(DB, "用户登录", "状态：成功 | 额外信息：用户ID{"+UserInfoRespon["user_id"]+"} Token: "+Token)
	ReturnMap["token"] = Token
	OutputLog(0, lib.JoinString("[User Auth] [", ConnectionTag, "] Auth Success, Username [", UserInfoRespon["username"], "] Get Token: {", Token, "}"))
	return HTTPJSONRespon("000", ReturnMap)
}

func AuthAppAuth(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ ParamMap map[string]string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(000) OK
	(010) Param `` Not Found
	(015) User Expired
	(016) User Disable
	(018) Token Invalid
	(019) User Access Denied
	(098) API Call Fail
	(099) Others Error
	ParamMap:
	@ token
	@ app_name
	ReturnMap:
	@ show_username string
	@ app_access bool
	*/
	if ParamMap["token"] == "" {
		OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] Param `token` Not Found"))
		return HTTPJSONRespon("010", "Param `token` Not Found")
	}
	if ParamMap["app_name"] == "" {
		OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] Param `app_name` Not Found"))
		return HTTPJSONRespon("010", "Param `app_name` Not Found")
	}
	RedisUserInfoKey, RedisUserInfoKeyCode, _ := lib.RedisKeys(RedisClient, CacheKeyGen("*", ParamMap["token"]))
	if RedisUserInfoKeyCode != 0 {
		OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] Token Not Found"))
		return HTTPJSONRespon("018", "Token Invalid")
	}
	UserInfoString, code, err := lib.RedisGet(RedisClient, RedisUserInfoKey[0])
	if code != 0 {
		if code == -26 {
			OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] Token Not Found"))
			return HTTPJSONRespon("018", "Token Invalid")
		}
		OutputLog(-2, lib.JoinString("[App Auth] [", ConnectionTag, "] Redis Get Fail: "+err.Error()))
		return HTTPJSONRespon("098", "API Call Fail")
	}
	UserInfo := make(map[string]string)
	err = json.Unmarshal([]byte(UserInfoString), &UserInfo)
	if err != nil {
		OutputLog(-2, lib.JoinString("[App Auth] [", ConnectionTag, "] String to Map Fail"))
		return HTTPJSONRespon("098", "API Call Fail")
	}
	if UserInfo["status"] != "true" {
		DataBaseLog(DB, "APP验证", "状态：失败 | 额外信息：用户ID{"+UserInfo["user_id"]+"}已禁用")
		OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] User Disable"))
		return HTTPJSONRespon("016", "User Disable")
	}
	if UserInfo["expire_time"] != "0" {
		if lib.StringNumCompare(UserInfo["expire_time"], lib.GetTimestamp_S_String()) < 0 {
			DataBaseLog(DB, "APP验证", "状态：失败 | 额外信息：用户ID{"+UserInfo["user_id"]+"}已过期 | 过期时间：["+lib.TimestampToDate(func(TimestampString string) int64 {
				t, _ := strconv.ParseInt(TimestampString, 10, 64)
				return t
			}(UserInfo["expire_time"]))+"]")
			OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] User Expired"))
			return HTTPJSONRespon("015", "User Expired")
		}
	}
	AccessAppTag := false
	var AppArray []string
	_ = json.Unmarshal([]byte(lib.Base64ToString(UserInfo["app"])), &AppArray)
	if IsAccess(AppArray, ParamMap["app_name"]) {
		AccessAppTag = true
	}
	ReturnMap := make(map[string]interface{})
	if AccessAppTag {
		ReturnMap["app_access"] = true
		ReturnMap["show_username"] = UserInfo["show_username"]
	} else {
		DataBaseLog(DB, "APP验证", "状态：失败 | 额外信息：用户ID{"+UserInfo["user_id"]+"}没有APP{"+ParamMap["app_name"]+"}的权限")
		OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] User Access Denied, AppName: {", ParamMap["app_name"], "}"))
		return HTTPJSONRespon("019", "User Access Denied")
	}
	DataBaseLog(DB, "APP验证", "状态：成功 | 额外信息：用户ID{"+UserInfo["user_id"]+"} APP: "+ParamMap["app_name"])
	OutputLog(0, lib.JoinString("[App Auth] [", ConnectionTag, "] Auth Success, Username [", UserInfo["username"], "] Auth Token: {", ParamMap["token"], "}, AppName: {", ParamMap["app_name"], "}"))
	return HTTPJSONRespon("000", ReturnMap)
}

func AuthTokenExist(ParamMap map[string]string, ConnectionTag string) string {
	/**
	Request Param:
	@ ParamMap map[string]string
	Respon Param:
	@ HTTPRespon string
	StatusCode:
	(000) OK
	(010) Param `` Not Found
	(018) Token Invalid
	(098) API Call Fail
	(099) Others Error
	ParamMap:
	@ token
	ReturnMap:
	*/
	if ParamMap["token"] == "" {
		OutputLog(0, lib.JoinString("[Token Exist] [", ConnectionTag, "] Param `token` Not Found"))
		return HTTPJSONRespon("010", "Param `token` Not Found")
	}
	RedisCheckToken, code, err := lib.RedisKeys(RedisClient, CacheKeyGen("*", ParamMap["token"]))
	if code != 0 {
		if code == -26 {
			OutputLog(0, lib.JoinString("[Token Exist] [", ConnectionTag, "] Token Not Found"))
			return HTTPJSONRespon("018", "Token Invalid")
		}
		OutputLog(-2, lib.JoinString("[Token Exist] [", ConnectionTag, "] Redis Get Fail: "+err.Error()))
		return HTTPJSONRespon("098", "API Call Fail")
	}
	if len(RedisCheckToken) <= 0 {
		OutputLog(0, lib.JoinString("[Token Exist] [", ConnectionTag, "] Token Not Found"))
		return HTTPJSONRespon("018", "Token Invalid")
	}
	OutputLog(0, lib.JoinString("[Token Exist] [", ConnectionTag, "] Token Found: ", ParamMap["token"]))
	return HTTPJSONRespon("000", "OK")
}
