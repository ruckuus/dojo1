package main

import (
	json2 "encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port           int            `json:"port"`
	Env            string         `json:"env"`
	HMACKey        string         `json:"hmac_key"`
	Pepper         string         `json:"pepper"`
	Database       PostgresConfig `json:"database"`
	Mailgun        MailgunConfig  `json:"mailgun"`
	RootPath       string         `json:"root_path"`
	AWSConfig      AWSConfig      `json:"aws_config"`
	StorageType    string         `json:"storage_type"`
	ImageCDNDomain string         `json:"image_cdn_domain"`
}

type MailgunConfig struct {
	APIKey       string `json:"api_key"`
	PublicAPIKey string `json:"public_api_key"`
	Domain       string `json:"domain"`
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:        3000,
		Env:         "dev",
		HMACKey:     "SuperSecret2019!$",
		Pepper:      "HALUSINOGEN2019$$",
		Database:    DefaultPostgresConfig(),
		RootPath:    "./",
		AWSConfig:   DefaultAWSConfig(),
		StorageType: "filesystem",
	}
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type AWSConfig struct {
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Region          string `json:"region"`
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

func DefaultAWSConfig() AWSConfig {
	return AWSConfig{
		Region:          "ap-southeast-1",
		AccessKeyID:     "XXXX",
		AccessKeySecret: "XXXX",
		Bucket:          "tataruma-images",
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
