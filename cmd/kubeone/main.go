package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/kubermatic/kubeone/pkg/command"
)

func main() {
	logger := setupLogging()

	app := cli.NewApp()
	app.Name = "kubeone"
	app.Usage = "Kubeone sets up Kubernetes clusters."
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "enable more verbose output",
		},
	}

	app.Commands = []cli.Command{
		command.InstallCommand(logger),
		command.ResetCommand(logger),
		command.KubeconfigCommand(logger),
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}

func setupLogging() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 MST",
	}

	return logger
}
