package main

import (
	"log"
	"os"

	"github.com/json-iterator/go/extra"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	flags    = make([]cli.Flag, 0)
	commands = make([]*cli.Command, 0)
)

func main() {
	extra.RegisterFuzzyDecoders()
	logrus.SetLevel(logrus.DebugLevel) // todo 改成可配置的

	app := &cli.App{
		Name:    "ca control",
		Usage:   "ca join cc and controlled by cc",
		Version: "1.0.0",
		Authors: []*cli.Author{
			{Name: "xiaobinqt", Email: "xiaobinqt@163.com"},
		},
		Flags:    flags,
		Commands: commands,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
