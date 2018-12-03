package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"
)

type controlPlane struct {
	ClusterName       string   `json:"cluster_name"`
	PublicAddress     []string `json:"public_address"`
	PrivateAddress    []string `json:"private_address"`
	Hostnames         []string `json:"hostnames"`
	SSHUser           string   `json:"ssh_user"`
	SSHPort           string   `json:"ssh_port"`
	SSHPrivateKeyFile string   `json:"ssh_private_key_file"`
	SSHAgentSocket    string   `json:"ssh_agent_socket"`
}

// Validate checks if the control plane conforms to our spec.
func (c *controlPlane) Validate() error {
	if len(c.PublicAddress) < 3 {
		return errors.New("must specify a unique cluster name")
	}

	if len(c.PublicAddress) != len(c.PrivateAddress) || len(c.PublicAddress) != len(c.Hostnames) {
		return errors.New("number of public addresses must be equal to number of private addresses and hostnames")
	}

	if len(c.PublicAddress) < 3 {
		return errors.New("must specify at least three public addresses")
	}

	for i := 0; i < len(c.PublicAddress); i++ {
		if len(c.PublicAddress[i]) == 0 {
			return fmt.Errorf("public address for host %d is empty", i+1)
		}

		if len(c.PrivateAddress[i]) == 0 {
			return fmt.Errorf("private address for host %d is empty", i+1)
		}

		if len(c.Hostnames[i]) == 0 {
			return fmt.Errorf("hostname for host %d is empty", i+1)
		}
	}

	return nil
}

type awsWorkerConfig struct {
	AMI                 string   `json:"ami"`
	AvailabilityZones   []string `json:"availability_zones"`
	DiskSize            string   `json:"disk_size"`
	DiskType            string   `json:"disk_type"`
	IAMInstanceProfile  string   `json:"iam_instance_profile"`
	InstanceType        string   `json:"instance_type"`
	Region              string   `json:"region"`
	SubnetID            string   `json:"subnet_id"`
	VPCID               string   `json:"vpc_id"`
	VPCSecurityGroupIDs []string `json:"vpc_security_group_ids"`
}

// Config represents configuration in the terraform output format
type Config struct {
	KubeOneAPI struct {
		Value struct {
			Endpoint string `json:"endpoint"`
		} `json:"value"`
	} `json:"kubeone_api"`

	KubeOneHosts struct {
		Value struct {
			ControlPlane []controlPlane `json:"control_plane"`
		} `json:"value"`
	} `json:"kubeone_hosts"`

	KubeOneWorkers struct {
		Value struct {
			AWS []awsWorkerConfig `json:"aws"`
		} `json:"value"`
	} `json:"kubeone_workers"`
}

// NewConfigFromJSON creates a new config object from json
func NewConfigFromJSON(j []byte) (c *Config, err error) {
	c = &Config{}
	return c, json.Unmarshal(j, c)
}

// Validate checks if the Terraform output conforms to our spec.
func (c *Config) Validate() error {
	planes := c.KubeOneHosts.Value.ControlPlane

	if len(planes) == 0 {
		return errors.New("no control plane specified")
	}

	if len(planes) > 1 {
		return errors.New("more than one control plane specified")
	}

	if err := planes[0].Validate(); err != nil {
		return fmt.Errorf("control plane is invalid: %v", err)
	}

	return nil
}

// Apply adds the terraform configuration options to the given
// cluster config.
func (c *Config) Apply(m *config.Cluster) {
	if c.KubeOneAPI.Value.Endpoint != "" {
		m.APIServer.Address = c.KubeOneAPI.Value.Endpoint
	}

	hosts := make([]config.HostConfig, 0)
	cp := c.KubeOneHosts.Value.ControlPlane[0]
	sshPort, _ := strconv.Atoi(cp.SSHPort)

	m.Name = cp.ClusterName

	// build up a list of master nodes
	for i, publicIP := range cp.PublicAddress {
		privateIP := publicIP
		if i < len(cp.PrivateAddress) {
			privateIP = cp.PrivateAddress[i]
		}

		// strip domain from hostname
		hostname := strings.Split(cp.Hostnames[i], ".")[0]

		hosts = append(hosts, config.HostConfig{
			ID:                i,
			PublicAddress:     publicIP,
			PrivateAddress:    privateIP,
			Hostname:          hostname,
			SSHUsername:       cp.SSHUser,
			SSHPort:           sshPort,
			SSHPrivateKeyFile: cp.SSHPrivateKeyFile,
			SSHAgentSocket:    cp.SSHAgentSocket,
		})
	}

	if len(hosts) > 0 {
		m.Hosts = hosts
	}

	// if there's a cloud provider specific configuration,
	// apply it to the worker nodes
	if len(c.KubeOneWorkers.Value.AWS) > 0 {
		aws := c.KubeOneWorkers.Value.AWS[0]

		az := ""
		if len(aws.AvailabilityZones) > 0 {
			az = aws.AvailabilityZones[0]
		}

		for idx, workerset := range m.Workers {
			setWorkersetFlag(&workerset, "ami", aws.AMI)
			setWorkersetFlag(&workerset, "availabilityZone", az)
			setWorkersetFlag(&workerset, "region", aws.Region)
			setWorkersetFlag(&workerset, "vpcId", aws.VPCID)
			setWorkersetFlag(&workerset, "subnetId", aws.SubnetID)
			setWorkersetFlag(&workerset, "instanceType", aws.InstanceType)
			setWorkersetFlag(&workerset, "instanceProfile", aws.IAMInstanceProfile)
			setWorkersetFlag(&workerset, "securityGroupIDs", aws.VPCSecurityGroupIDs)
			setWorkersetFlag(&workerset, "diskSize", aws.DiskSize)
			setWorkersetFlag(&workerset, "diskType", aws.DiskType)

			m.Workers[idx] = workerset
		}
	}
}

func setWorkersetFlag(w *config.WorkerConfig, name string, value interface{}) {
	// ignore empty values (i.e. not set in terraform output)
	if s, ok := value.(string); ok && s == "" {
		return
	}

	if slice, ok := value.([]string); ok && len(slice) == 0 {
		return
	}

	if _, exists := w.Spec[name]; !exists {
		w.Spec[name] = value
	}
}
