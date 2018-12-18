package config

import "errors"

// WorkerConfig describes a set of worker machines.
type WorkerConfig struct {
	Name     string         `json:"name"`
	Replicas int            `json:"replicas"`
	Config   providerConfig `json:"config"`
}

// Validate checks if the Config makes sense.
func (m *WorkerConfig) Validate() error {
	if m.Name == "" {
		return errors.New("no name given")
	}

	if m.Replicas < 1 {
		return errors.New("replicas must be >= 1")
	}

	return nil
}
