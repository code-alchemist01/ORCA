package container

import "time"

// ContainerSpec defines the specification for a container
type ContainerSpec struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Ports       map[string]string `json:"ports,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Volumes     []VolumeMount     `json:"volumes,omitempty"`
}

// VolumeMount defines a volume mount
type VolumeMount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ReadOnly    bool   `json:"read_only,omitempty"`
}

// Container represents a running container
type Container struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	Ports       map[string]string `json:"ports,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Created     time.Time         `json:"created"`
	Started     *time.Time        `json:"started,omitempty"`
}

// DeploymentSpec defines the specification for a deployment
type DeploymentSpec struct {
	Name      string        `json:"name"`
	Replicas  int           `json:"replicas"`
	Container ContainerSpec `json:"container"`
	Strategy  string        `json:"strategy,omitempty"`
}

// ServiceSpec defines the specification for a service
type ServiceSpec struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Selector map[string]string `json:"selector"`
	Ports    []ServicePort     `json:"ports"`
}

// ServicePort defines a service port mapping
type ServicePort struct {
	Port       int `json:"port"`
	TargetPort int `json:"target_port"`
}