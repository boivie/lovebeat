package config

import (
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"os"
)

type Config struct {
	Mail ConfigMail
}

type ConfigMail struct {
	From   string
	Server string
}

var log = logging.MustGetLogger("lovebeat")

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func ReadConfig(fname string) Config {
	var conf = Config{Mail: ConfigMail{From: "lovebeat@example.com",
		Server: "localhost:25"}}
	if e, _ := exists(fname); e {
		if _, err := toml.DecodeFile(fname, &conf); err != nil {
			log.Error("Failed to parse configuration file", err)
		}
	}
	return conf
}
