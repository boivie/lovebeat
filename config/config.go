package config

import (
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Mail     ConfigMail
	Slack    ConfigSlack
	Udp      ConfigBind
	Tcp      ConfigBind
	Http     ConfigBind
	Database ConfigDatabase
	Metrics  ConfigMetrics
	Alerts   map[string]ConfigAlert
	Views    map[string]ConfigView
	Eventlog ConfigEventlog
}

type ConfigMail struct {
	From   string
	Server string
}

type ConfigSlack struct {
	WebhookUrl string `toml:"webhook_url"`
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
	Mail         string
	Webhook      string
	SlackChannel string `toml:"slack_channel"`
	Script       string
}

type ConfigView struct {
	Regexp string
	Alerts []string
}

type ConfigEventlog struct {
	Path string
	Mode os.FileMode
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
		Slack: ConfigSlack{
			WebhookUrl: "https://hooks.slack.com/services/TXXXXXXXX/BXXXXXXXX/XXXXXXXXXXXXXXXXX",
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
		Eventlog: ConfigEventlog{
			Path: "",
			Mode: 644, // Reinterpreted as octal below
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
	mode, _ := strconv.ParseInt(strconv.FormatInt(int64(conf.Eventlog.Mode), 10), 8, 64)
	conf.Eventlog.Mode = os.FileMode(mode)
	return conf
}
