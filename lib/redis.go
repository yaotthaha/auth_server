package lib

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

func RedisConnect(ConfigMap map[string]string) (*redis.Client, int, error) {
	/**
	Request Param:
	@ ConfigMap map[string]string
	Respon Param:
	@ RedisClient *redis.Client
	@ StatusCode int
	@ Error error
	StatusCode:
	(0) OK
	(-20) Redis Connect Fail
	*/
	RedisOptions := &redis.Options{}
	RedisConnArray := strings.Split(ConfigMap["url"], "://")
	switch RedisConnArray[0] {
	case "tcp":
		RedisOptions.Network = "tcp"
	case "unix":
		RedisOptions.Network = "unix"
	}
	RedisOptions.Addr = RedisConnArray[1]
	RedisOptions.DB = 0
	RedisOptions.Password = ConfigMap["pass"]
	RedisClient := redis.NewClient(RedisOptions)
	var timeout int64 = 3
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	err := RedisClient.Ping(ctx).Err()
	if err != nil {
		return nil, -20, err
	}
	return RedisClient, 0, nil
}

func RedisConnClose(RedisClient *redis.Client) (int, error) {
	/**
	Request Param:
	@ RedisClient *redis.Client
	Respon Param:
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-21) Redis Connection Close Fail
	*/
	err := RedisClient.Close()
	if err != nil {
		return -21, err
	}
	return 0, nil
}

func RedisSet(RedisClient *redis.Client, Key string, Value string, ExpireTime int64) (int, error) {
	/**
	Request Param:
	@ RedisClient *redis.Client
	@ Key string
	@ Value string
	@ ExpireTime int64
	Respon Param:
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-22) Redis Set Fail
	*/
	err := RedisClient.Set(context.Background(), Key, Value, time.Second*time.Duration(ExpireTime)).Err()
	if err != nil {
		return -22, err
	}
	return 0, nil
}

func RedisGet(RedisClient *redis.Client, Key string) (string, int, error) {
	/**
	Request Param:
	@ RedisClient *redis.Client
	@ Key string
	Respon Param:
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-23) Redis Get Fail
	(-26) Redis Key Not Found
	*/
	Result, err := RedisClient.Get(context.Background(), Key).Result()
	if err == redis.Nil {
		return "", -26, err
	} else if err != nil {
		return "", -23, err
	}
	return Result, 0, nil
}

func RedisExpire(RedisClient *redis.Client, Key string, ExpireTime int64) (int, error) {
	/**
	Request Param:
	@ RedisClient *redis.Client
	@ Key string
	@ ExpireTime int64
	Respon Param:
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-24) Redis Expire Fail
	*/
	err := RedisClient.ExpireAt(context.Background(), Key, time.Now().Add(time.Second*time.Duration(ExpireTime))).Err()
	if err != nil {
		return -24, err
	}
	return 0, nil
}

func RedisDel(RedisClient *redis.Client, Key string) (int, error) {
	/**
	Request Param:
	@ RedisClient *redis.Client
	@ Key string
	Respon Param:
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-25) Redis Del Fail
	*/
	err := RedisClient.Del(context.Background(), Key).Err()
	if err != nil {
		return -25, err
	}
	return 0, nil
}

func RedisKeys(RedisClient *redis.Client, KeyPattern string) ([]string, int, error) {
	/**
	Request Param:
	@ RedisClient *redis.Client
	@ KeyPattern string
	Respon Param:
	@ Keys []string
	@ StatusCode
	@ Error error
	StatusCode:
	(0) OK
	(-26) Redis Key Not Found
	*/
	Result := RedisClient.Keys(context.Background(), KeyPattern)
	if len(Result.Val()) <= 0 {
		return nil, -26, errors.New("Redis Key Not Found")
	}
	return Result.Val(), 0, nil
}
