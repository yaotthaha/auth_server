package main

import (
	"auth_server/lib"
	"database/sql"
	"errors"
)

func DataBaseSearchUserInfo(DB *sql.DB, UserID string) (map[string]string, int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ UserID string
	Respon Param:
	@ UserInfo map[string]string
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-50) UserInfo Error
	(-51) User Not Found
	(...)
	*/
	SQLSearchUserInfo := "SELECT * FROM `" + DB_TableName + "` WHERE `user_id` = '" + UserID + "'"
	UserInfoAll, code, err := lib.DataBaseQuery(DB, SQLSearchUserInfo)
	if code != 0 {
		return nil, code, err
	}
	if UserInfoAll == nil {
		return nil, 0, nil
	}
	if len(UserInfoAll) > 1 {
		return nil, -50, errors.New("UserInfo Error")
	}
	if len(UserInfoAll) == 0 {
		return nil, -51, errors.New("User Not Found")
	}
	return UserInfoAll[0], 0, nil
}
