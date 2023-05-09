package managers

import (
	"flag"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

var environmentLog = logrus.WithField("fun", "environment")

const (
	TOKEN = "T"
)

var (
	CONFIG     Config
	configFile string
)

type Config struct {
	Version   string      `toml:"version"`
	Env       string      `toml:"env"`
	Port      int         `toml:"port"`
	ServerURL string      `toml:"serverURL"`
	DB        DBConfig    `toml:"mongodb"`
	Redis     RedisConfig `toml:"redis"`
}

type DBConfig struct {
	URI      string `toml:"URI"`
	Database string `toml:"Database"`
}

type RedisConfig struct {
	URL      string `toml:"URL"`
	Password string `toml:"Password"`
	DB       int    `toml:"Database"`
}

func init() {
	flag.StringVar(&configFile, "c", "config.toml", "config file of CONFIG ")
}

func Environment() {
	//加载配置文件
	flag.Parse()
	file, err := os.Open(configFile)
	if err != nil {
		environmentLog.WithError(err).Panic("Read config.toml failed")
	}
	defer file.Close()
	if err = toml.NewDecoder(file).Decode(&CONFIG); err != nil {
		environmentLog.WithError(err).Panic("Decode config failed")
	}
	environmentLog.Infof("Environment: %s", CONFIG.Env)
}
