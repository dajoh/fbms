package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ExpireTime     int
	ListenAddress  string
	MaxPayloadSize int
}

func LoadConfig(filename string) *Config {
	conf := new(Config)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
