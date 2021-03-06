package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	DB struct {
		UserName string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"db"`
	JWTSecret string `json:"jwtSecret"`
}

var ErrFileNotExists = os.ErrNotExist

func ParseFromFile(fileName string) (Config, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, ErrFileNotExists
		}
		return Config{}, err
	}
	defer f.Close()

	c := Config{}

	d := json.NewDecoder(f)
	err = d.Decode(&c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (c Config) MakeDBString() string {
	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
	        c.DB.Host, c.DB.UserName, c.DB.Password, c.DB.Database)
	return psqlInfo
}

var defaultJSON = []byte(`{
	"db": {
		"username": "",
		"password": "",
		"database": "frengine",
		"host": "localhost",
		"port": 3306
	},
	"jwtSecret": "secret for generating JWT keys here"
}
`)

func WriteDefault(fileName string) error {
	return ioutil.WriteFile(fileName, defaultJSON, 0600)
}
