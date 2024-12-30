package config
import (
    "fmt"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
}



func NewDBConfig() *DBConfig {
    return &DBConfig {
        Host:     "localhost",
        Port:     9030,
        User:     "root",
        Password: "root",
        Database: "dw_monitor",
    }
}


func (c *DBConfig) GetDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", c.User, c.Password, c.Host, c.Port, c.Database)
}