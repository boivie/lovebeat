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
	General  ConfigGeneral
	Mail     ConfigMail
	Slack    ConfigSlack
	Udp      ConfigBind
	Tcp      ConfigBind
	Http     ConfigBind
	Database ConfigDatabase
	Metrics  ConfigMetrics
	Alarms   []ConfigAlarm
	Alerts   map[string]ConfigAlert
	Eventlog ConfigEventlog
	Notify   []ConfigNotify
}

type ConfigGeneral struct {
	PublicUrl string `toml:"public_url"`
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

	RemoteS3Url    string `toml:"remote_s3_url"`
	RemoteS3Region string `toml:"remote_s3_region"`
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

type ConfigAlarm struct {
	Name     string
	Pattern  string
	Includes []string
	Excludes []string
	Alerts   []string
}

type ConfigEventlog struct {
	Path string
	Mode os.FileMode
}

type ConfigNotify struct {
	Lovebeat string
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
		log.Infof("Reading configuration file %s", fname)
		if _, err := toml.DecodeFile(fname, conf); err != nil {
			log.Errorf("Failed to parse configuration file '%s': %v", fname, err)
		}
	}
}

func ReadConfig(fname string, dirname string) Config {
	var conf = Config{
		General: ConfigGeneral{
			PublicUrl: "http://lovebeat.example.com/",
		},
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
