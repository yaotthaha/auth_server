package lib

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

func DataBaseConnect(ConfigMap map[string]string) (*sql.DB, int, error) {
	/**
	Request Param:
	Respon Param:
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-16) [Server] DataBase Connect Fail
	*/
	DataBaseConnArray := strings.Split(ConfigMap["url"], "://")
	switch DataBaseConnArray[0] {
	case "tcp", "unix":
		var connect_url string
		switch DataBaseConnArray[0] {
		case "tcp":
			connect_url = ConfigMap["user"] + ":" + ConfigMap["pass"] + "@tcp(" + DataBaseConnArray[1] + ")/" + ConfigMap["db_name"]
		case "unix":
			connect_url = ConfigMap["user"] + ":" + ConfigMap["pass"] + "@unix(" + DataBaseConnArray[1] + ")/" + ConfigMap["db_name"]
		}
		DB, err := sql.Open("mysql", connect_url)
		DB.SetConnMaxLifetime(16)
		DB.SetMaxIdleConns(8)
		if err != nil {
			return nil, -16, err
		}
		err = DB.Ping()
		if err != nil {
			return nil, -16, err
		}
		return DB, 0, nil
	case "file":
		DB, err := sql.Open("sqlite3", DataBaseConnArray[1])
		if err != nil {
			return nil, -16, err
		}
		return DB, 0, nil
	default:
		return nil, -15, errors.New("Param `DB_ConnectionURL` Invalid")
	}
}

func DataBaseCloseConn(DB *sql.DB) (int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	Respon Param:
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-17) [Server] DataBase Connection Close Fail
	*/
	err := DB.Close()
	if err != nil {
		return -17, errors.New("DataBase Connection Close Fail")
	}
	return 0, nil
}

func DataBaseQuery(DB *sql.DB, SQL string) ([]map[string]string, int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ SQL string
	Respon Param:
	@ DataMap []map[string]interface{}
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-18) SQL Query Fail
	*/
	rows, err := DB.Query(SQL)
	if err != nil {
		return nil, -18, err
	}
	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for index, _ := range cache {
		var a interface{}
		cache[index] = &a
	}
	var List []map[string]string
	for rows.Next() {
		_ = rows.Scan(cache...)
		item := make(map[string]string)
		for i, data := range cache {
			temp := *data.(*interface{})
			item[columns[i]] = string(temp.([]uint8))
		}
		List = append(List, item)
	}
	_ = rows.Close()
	return List, 0, nil
}

func DataBaseExec(DB *sql.DB, SQL string) (int, error) {
	/**
	Request Param:
	@ DB *sql.DB
	@ SQL string
	Respon Param:
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-19) SQL Exec Fail
	*/
	_, err := DB.Exec(SQL)
	if err != nil {
		return -19, err
	}
	return 0, nil
}
