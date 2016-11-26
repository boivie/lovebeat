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
	Mail     ConfigSmtp
	Mailgun  ConfigMailgun
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

type ConfigSmtp struct {
	From   string
	Server string
}

type ConfigMailgun struct {
	From   string
	Domain string
	ApiKey string `toml:"api_key"`
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

func readFile(conf *Config, fname string) error {
	if e, _ := exists(fname); e {
		log.Infof("Reading configuration file %s", fname)
		if _, err := toml.DecodeFile(fname, conf); err != nil {
			log.Errorf("Failed to parse configuration file '%s': %v", fname, err)
			return err
		}
	}
	return nil
}

func ReadConfig(fname string, dirname string) (conf Config, err error) {
	conf = Config{
		General: ConfigGeneral{
			PublicUrl: "http://lovebeat.example.com/",
		},
		Mail: ConfigSmtp{
			From:   "lovebeat@example.com",
			Server: "localhost:25",
		},
		Mailgun: ConfigMailgun{},
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
	err = readFile(&conf, fname)
	if err != nil {
		return
	}
	if dirname != "" {
		var files []os.FileInfo
		files, err = ioutil.ReadDir(dirname)
		if err == nil {
			for _, f := range files {
				path := filepath.Join(dirname, f.Name())
				err = readFile(&conf, path)
				if err != nil {
					return
				}
			}
		} else {
			// This directory is optional
			err = nil
		}
	}
	mode, _ := strconv.ParseInt(strconv.FormatInt(int64(conf.Eventlog.Mode), 10), 8, 64)
	conf.Eventlog.Mode = os.FileMode(mode)
	return
}
