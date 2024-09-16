package config

type Config struct {
	ServerPort       string
	DockerAPIVersion string
	BadgerDBPath     string
}

func NewConfig() *Config {
	return &Config{
		ServerPort:       ":8080",
		DockerAPIVersion: "1.41",
		BadgerDBPath:     "./badger",
	}
}
