package db

import (
	"errors"
	"github.com/gomodule/redigo/redis"
)

type Cache interface {
	SetSession(string, string) error
	GetSession(string) (string, error)
	SetData(string, string, string) error
	GetData(string, string) (string, error)
}

type RedisCache struct {
	Conn redis.Conn
}

func (r *RedisCache) SetSession(user string, sessionid string) error {
	_, err := r.Conn.Do("set", sessionid, user, "EX", 120)
	return err
}

func (r *RedisCache) GetSession(sessionid string) (string, error) {
	reply, err := r.Conn.Do("get", sessionid)
	if err != nil {
		return "", err
	}
	if reply == nil {
		return "", errors.New("no exits")
	}
	return string(reply.([]uint8)), nil
}

func (r *RedisCache) SetData(user string, key string, value string) error {
	_, err := r.Conn.Do("hset", user, key, value)
	return err
}

func (r *RedisCache) GetData(user string, key string) (string, error) {
	reply, err := r.Conn.Do("hget", user, key)
	if err != nil {
		return "", err
	} else {
		if reply == nil {
			return "", errors.New("no value")
		}

		return string(reply.([]uint8)), nil
	}
}
