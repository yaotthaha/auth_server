package main

import (
	"auth_server/lib"
	"database/sql"
	"encoding/json"
	"strconv"
)

func DataBaseCheckFieldExist(DB *sql.DB, Field string, Value string) (bool, int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ Field string
	@ Value string
	Respon Param:
	@ Check bool
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(...)
	*/
	SQLSearchField := "SELECT `" + Field + "` FROM `" + DB_TableName + "` WHERE `" + Field + "` = '" + Value + "'"
	Result, code, err := lib.DataBaseQuery(DB, SQLSearchField)
	if code != 0 {
		return false, code, err
	}
	if len(Result) < 1 {
		return false, 0, nil
	}
	return true, 0, nil
}

func DataBaseUserAdd(DB *sql.DB, user_id, username, status, password_secret, app, expire_time, show_username string) (int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ user_id string
	@ username string
	@ status string
	@ password_secret string
	@ app string
	@ expire_time string
	@ show_username string
	Respon Param:
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-19) SQL Exec Fail
	*/
	SQLUserAdd := "INSERT INTO `" + DB_TableName + "` (" +
		"`uuid`, " +
		"`username`, " +
		"`user_id`, " +
		"`status`, " +
		"`password_secret`, " +
		"`app`, " +
		"`expire_time`, " +
		"`show_username`" +
		") VALUES (" +
		"'" + lib.GenUUID() + "'," +
		"'" + username + "'," +
		"'" + user_id + "'," +
		"'" + status + "'," +
		"'" + password_secret + "'," +
		"'" + app + "'," +
		"'" + expire_time + "'," +
		"'" + show_username + "'" +
		")"
	code, err := lib.DataBaseExec(DB, SQLUserAdd)
	if code != 0 {
		return code, err
	}
	return 0, nil
}

func DataBaseUserDel(DB *sql.DB, uuid_index string) (int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ user_id string
	Respon Param:
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-19) SQL Exec Fail
	*/
	SQLUserDel := "DELETE FROM `" + DB_TableName + "` WHERE `uuid` = '" + uuid_index + "'"
	code, err := lib.DataBaseExec(DB, SQLUserDel)
	if code != 0 {
		return code, err
	}
	return 0, nil
}

func DataBaseUserSet(DB *sql.DB, uuidIndex string, InputMap map[string]string) (int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ InputMap mpa[string]string
	Respon Param:
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-19) SQL Exec Fail
	*/
	SqlUserSet1 := "UPDATE `" + DB_TableName + "` SET "
	SqlUserSet2 := ""
	SqlUserSet3 := "WHERE `uuid` = '" + uuidIndex + "'"
	for k, v := range InputMap {
		SqlUserSet2 += "`" + k + "` = '" + v + "', "
	}
	SqlUserSet2 = SqlUserSet2[0:len(SqlUserSet2)-2] + " "
	SQLUserSet := SqlUserSet1 + SqlUserSet2 + SqlUserSet3
	code, err := lib.DataBaseExec(DB, SQLUserSet)
	if code != 0 {
		return code, err
	}
	return 0, nil
}

func DataBaseUserList(DB *sql.DB) ([]interface{}, int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	Respon Param:
	@ ResponMap []interface{}
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-18) SQL Query Fail
	*/
	Result, code, err := lib.DataBaseQuery(DB, "SELECT * FROM `"+DB_TableName+"`")
	if code != 0 {
		return nil, code, err
	}
	ResultReal := make([]interface{}, len(Result))
	for k1, v1 := range Result {
		TempMap := make(map[string]interface{})
		for k2, v2 := range v1 {
			if k2 == "uuid" {
				continue
			} else if k2 == "status" {
				switch v2 {
				case "true":
					TempMap[k2] = true
				case "false":
					TempMap[k2] = false
				}
			} else if k2 == "password_secret" {
				continue
			} else if k2 == "expire_time" {
				if v2 == "0" {
					TempMap[k2] = "Permanent"
				} else {
					Num, _ := strconv.ParseInt(v2, 10, 64)
					TempMap[k2] = lib.TimestampToDate(Num)
				}
			} else if k2 == "app" {
				AppArrayString := lib.Base64ToString(v2)
				if AppArrayString == `["all"]` {
					TempMap[k2] = "all"
				} else {
					var AppArray []string
					_ = json.Unmarshal([]byte(AppArrayString), &AppArray)
					TempMap[k2] = AppArray
				}
			} else {
				TempMap[k2] = v2
			}
		}
		ResultReal[k1] = TempMap
	}
	return ResultReal, 0, nil
}
