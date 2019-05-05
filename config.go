package main

import "fmt"

type Config struct {
	Port    int    `json:"port"`
	Env     string `json:"env"`
	HMACKey string `json:"hmac_key"`
	Pepper  string `json:"pepper"`
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:    3000,
		Env:     "dev",
		HMACKey: "SuperSecret2019!$",
		Pepper:  "HALUSINOGEN2019$$",
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
