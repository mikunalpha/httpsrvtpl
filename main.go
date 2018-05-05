package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/mikunalpha/httpsrvtpl/store"
	"github.com/mikunalpha/httpsrvtpl/store/mock"

	"github.com/mikunalpha/httpsrvtpl/server"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var version = "v0.1.2"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "address",
		Value:  "0.0.0.0:8888",
		Usage:  "server will listen on the address",
		EnvVar: "_ADDRESS",
	},
	cli.StringFlag{
		Name:   "database-type",
		Value:  "mock",
		Usage:  "database type",
		EnvVar: "_DATABASE_TYPE",
	},
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "show debug message",
		EnvVar: "_DEBUG",
	},
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func newStore(c *cli.Context) (store.Store, error) {
	switch c.GlobalString("database-type") {
	case "mock":
		return mock.New(), nil
	}
	return nil, fmt.Errorf("unknow database type %s", c.GlobalString("database-type"))
}

func action(c *cli.Context) error {
	if c.GlobalIsSet("debug") {
		log.SetLevel(log.DebugLevel)
	}

	st, err := newStore(c)
	if err != nil {
		return fmt.Errorf("newStore failed: %v", err)
	}

	opts := []server.Option{
		server.OptStore(st),
		server.OptAllowMethodOverride(),
		server.OptAddPingHandler(),
		server.OptAddDebugHandler(),
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	s := server.New(c.GlobalString("address"), opts...)
	s.Run()

	<-stop

	return s.Stop()
}

func main() {
	app := cli.NewApp()
	app.Name = "httpsrvtpl"
	app.Usage = "HTTPSRVTPL IS AWESOME"
	app.UsageText = "httpsrvtpl [options]"
	app.Version = version
	app.Copyright = "(c) 2018 mikun800527@gmail.com"
	app.HideHelp = true
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelp(c)
		return nil
	}
	app.Flags = flags
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
	}
}
