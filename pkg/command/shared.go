package command

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func handleErrors(logger logrus.FieldLogger, action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		err := action(ctx)
		if err != nil {
			logger.Errorln(err)
			err = cli.NewExitError("", 1)
		}

		return err
	}
}

func setupLogger(logger *logrus.Logger, action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		if ctx.GlobalBool("verbose") {
			logger.SetLevel(logrus.DebugLevel)
		}

		return action(ctx)
	}
}
