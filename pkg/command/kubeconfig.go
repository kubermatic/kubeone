package command

import (
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// KubeconfigCommand returns the structure for declaring the "install" subcommand.
func KubeconfigCommand(logger *logrus.Logger) cli.Command {
	return cli.Command{
		Name:      "kubeconfig",
		Usage:     "Prints the kubeconfig for a successfully installed cluster",
		ArgsUsage: "CLUSTER_FILE",
		Action:    KubeconfigAction(logger),
		Flags: []cli.Flag{
			cli.StringFlag{
				EnvVar: "TF_OUTPUT",
				Name:   "tfjson, t",
				Usage:  "path to terraform output JSON or - for stdin",
				Value:  "",
			},
		},
	}
}

// KubeconfigAction wrapper for logger
func KubeconfigAction(logger *logrus.Logger) cli.ActionFunc {
	return handleErrors(logger, setupLogger(logger, func(ctx *cli.Context) error {
		// load cluster config
		clusterFile := ctx.Args().First()
		if clusterFile == "" {
			return errors.New("no cluster config file given")
		}

		cluster, err := loadClusterConfig(clusterFile)
		if err != nil {
			return fmt.Errorf("failed to load cluster: %v", err)
		}

		// apply terraform
		tf := ctx.String("tfjson")
		if err = applyTerraform(tf, cluster); err != nil {
			return err
		}

		if err = cluster.DefaultAndValidate(); err != nil {
			return err
		}

		// connect to leader
		leader := cluster.Leader()
		connector := ssh.NewConnector()

		conn, err := connector.Connect(*leader)
		if err != nil {
			return fmt.Errorf("failed to connect to leader: %v", err)
		}
		defer conn.Close()

		// get the kubeconfig
		kubeconfig, _, _, err := conn.Exec("sudo cat /etc/kubernetes/admin.conf")
		if err != nil {
			return fmt.Errorf("failed to read kubeconfig: %v", err)
		}

		fmt.Println(kubeconfig)

		return nil
	}))
}
