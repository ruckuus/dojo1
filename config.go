package main

import (
	json2 "encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	HMACKey  string         `json:"hmac_key"`
	Pepper   string         `json:"pepper"`
	Database PostgresConfig `json:"database"`
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		HMACKey:  "SuperSecret2019!$",
		Pepper:   "HALUSINOGEN2019$$",
		Database: DefaultPostgresConfig(),
	}
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (p PostgresConfig) Dialect() string {
	return "postgres"
}

func (p PostgresConfig) ConnectionInfo() string {
	if p.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			p.Host, p.Port, p.User, p.Name)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Name)

}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     54320,
		User:     "lenslocked_user",
		Password: "lenslocked_password",
		Name:     "lenslocked_db",
	}
}

func LoadConfig(isProd bool) Config {
	if !isProd {
		DefaultConfig()
	}

	cfgFile, err := os.Open(".config")
	if err != nil {
		panic(err)
	}

	json := json2.NewDecoder(cfgFile)

	var cfg Config

	err = json.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
