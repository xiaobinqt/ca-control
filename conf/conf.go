package conf

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Conf struct {
	App App    `json:"app" yaml:"app"`
	Db  DbConf `json:"db" yaml:"db"`
}

type App struct {
	Env      string `json:"env" yaml:"env"`
	LogLevel string `json:"log_level" yaml:"log_level"`
}

type DbConf struct {
	IP      string `json:"ip" yaml:"ip"`
	Port    string `json:"port" yaml:"port"`
	DbName  string `json:"db_name" yaml:"db_name"`
	User    string `json:"user" yaml:"user"`
	Passwd  string `json:"passwd" yaml:"passwd"`
	Charset string `json:"charset" yaml:"charset"`
}

// GetConf .
func GetConf(cfg string) (conf *Conf, err error) {
	var (
		yamlFile = make([]byte, 0)
	)

	filepath := fmt.Sprintf("%sconfig.yaml", cfg)
	logrus.Infof("filepath: %s", filepath)
	yamlFile, err = ioutil.ReadFile(filepath)
	if err != nil {
		err = errors.Wrapf(err, "ReadFile error")
		logrus.Errorf(err.Error())
		return conf, err
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		err = errors.Wrapf(err, "yaml.Unmarshal error")
		logrus.Errorf(err.Error())
		return conf, err
	}

	return conf, nil
}

// GetDbConf .
func GetDbConf(conf *Conf) DbConf {
	return conf.Db
}
