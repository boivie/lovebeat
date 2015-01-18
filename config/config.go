package config

import (
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"os"
)

type Config struct {
	Mail     ConfigMail
	Udp      ConfigBind
	Tcp      ConfigBind
	Http     ConfigBind
	Database ConfigDatabase
}

type ConfigMail struct {
	From   string
	Server string
}

type ConfigBind struct {
	Listen string
}
type ConfigDatabase struct {
	Filename string
	Interval int
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
	var conf = Config{
		Mail: ConfigMail{
			From:   "lovebeat@example.com",
			Server: "localhost:25",
		},
		Udp: ConfigBind{
			Listen: ":8127",
		},
		Tcp: ConfigBind{
			Listen: ":8127",
		},
		Http: ConfigBind{
			Listen: ":8080",
		},
		Database: ConfigDatabase{
			Filename: "lovebeat-data.gz",
			Interval: 60,
		},
	}
	if e, _ := exists(fname); e {
		if _, err := toml.DecodeFile(fname, &conf); err != nil {
			log.Error("Failed to parse configuration file", err)
		}
	}
	return conf
}
