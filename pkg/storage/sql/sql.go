package sql

import (
	"database/sql"
	"fmt"
)

type Config struct {
	Username string
	Password string
	Host     string
	DB       string
	Prefix   string // table prefix, default is nd_
	Port     int
	MaxIdle  int
	MaxOpen  int
	ShowSQL  bool
}

func (c *Config) Dns() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=Local&parseTime=true",
		c.Username, c.Password, c.Host, c.Port, c.DB)
}

// check db
func (c *Config) CheckDB() error {
	// 先确认数据库可以正常链接
	where := fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.Username, c.Password, c.Host, c.Port)
	db, err := sql.Open("mysql", where)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return err
	}
	// 检查要使用的数据库是否存在，不存在创建
	where = fmt.Sprintf("%s:%s@tcp(%s:%d)/INFORMATION_SCHEMA?charset=utf8mb4", c.Username, c.Password, c.Host, c.Port)
	connection, err := sql.Open("mysql", where)
	if err != nil {
		return err
	}
	defer connection.Close()
	_, err = connection.Exec(fmt.Sprintf("create database if not exists %s;", c.DB))
	if err != nil {
		return err
	}

	return nil
}
