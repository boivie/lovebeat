package config

import (
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Mail     ConfigMail
	Slackhook     ConfigSlackhook
	Udp      ConfigBind
	Tcp      ConfigBind
	Http     ConfigBind
	Database ConfigDatabase
	Metrics  ConfigMetrics
	Alerts   map[string]ConfigAlert
	Views    map[string]ConfigView
}

type ConfigMail struct {
	From   string
	Server string
}

type ConfigSlackhook struct {
	Uri   string
	Template string
}

type ConfigBind struct {
	Listen string
}

type ConfigDatabase struct {
	Filename string
	Interval int
}

type ConfigMetrics struct {
	Server string
	Prefix string
}

type ConfigAlert struct {
	Mail    string
	Webhook string
	Slackhook string
}

type ConfigView struct {
	Regexp string
	Alerts []string
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

func readFile(conf *Config, fname string) {
	if e, _ := exists(fname); e {
		log.Info("Reading configuration file %s", fname)
		if _, err := toml.DecodeFile(fname, conf); err != nil {
			log.Error("Failed to parse configuration file %s", fname, err)
		}
	}
}

func ReadConfig(fname string, dirname string) Config {
	var conf = Config{
		Mail: ConfigMail{
			From:   "lovebeat@example.com",
			Server: "localhost:25",
		},
		Slackhook: ConfigSlackhook{
			Uri:   "http://localhost:1323/services/deded/wswswsw/erferferferferf",
			Template: `{"channel": "#alert-test", "username": "webhookbot", "text": "slack check {{.View.Name}}-{{.View.IncidentNbr}}", "icon_emoji": ":ghost:"}`,
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
			Filename: "lovebeat.db",
			Interval: 60,
		},
		Metrics: ConfigMetrics{
			Server: "",
			Prefix: "lovebeat",
		},
	}
	readFile(&conf, fname)
	if dirname != "" {
		files, err := ioutil.ReadDir(dirname)
		if err == nil {
			for _, f := range files {
				path := filepath.Join(dirname, f.Name())
				readFile(&conf, path)
			}
		}
	}
	return conf
}
