package container

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

// Manager handles container operations
type Manager struct {
	client *client.Client
	logger *logrus.Logger
}

// NewManager creates a new container manager
func NewManager(logger *logrus.Logger) (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client oluşturulamadı: %w", err)
	}

	return &Manager{
		client: cli,
		logger: logger,
	}, nil
}

// Create creates a new container from spec
func (m *Manager) Create(ctx context.Context, spec ContainerSpec) (*Container, error) {
	// Port bindings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for containerPort, hostPort := range spec.Ports {
		// Parse container port (remove protocol if present)
		portStr := containerPort
		protocol := "tcp"
		if strings.Contains(containerPort, "/") {
			parts := strings.Split(containerPort, "/")
			if len(parts) == 2 {
				portStr = parts[0]
				protocol = parts[1]
			}
		}

		port, err := nat.NewPort(protocol, portStr)
		if err != nil {
			return nil, fmt.Errorf("geçersiz port: %s", containerPort)
		}

		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	// Environment variables
	env := make([]string, 0, len(spec.Environment))
	for key, value := range spec.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Container config
	config := &container.Config{
		Image:        spec.Image,
		Env:          env,
		Labels:       spec.Labels,
		ExposedPorts: exposedPorts,
		WorkingDir:   spec.WorkingDir,
	}

	if len(spec.Command) > 0 {
		config.Cmd = spec.Command
	}

	// Host config
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}

	// Network config
	networkConfig := &network.NetworkingConfig{}

	// Create container
	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, spec.Name)
	if err != nil {
		return nil, fmt.Errorf("docker container oluşturulamadı: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"container_id": resp.ID,
		"name":         spec.Name,
		"image":        spec.Image,
	}).Info("Container oluşturuldu")

	return &Container{
		ID:          resp.ID,
		Name:        spec.Name,
		Image:       spec.Image,
		Status:      "created",
		Ports:       spec.Ports,
		Environment: spec.Environment,
		Labels:      spec.Labels,
		Created:     time.Now(),
	}, nil
}

// Start starts a container
func (m *Manager) Start(ctx context.Context, containerID string) error {
	err := m.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("container başlatılamadı: %w", err)
	}

	m.logger.WithField("container_id", containerID).Info("Container başlatıldı")
	return nil
}

// Stop stops a container
func (m *Manager) Stop(ctx context.Context, containerID string) error {
	timeout := 30
	err := m.client.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("container durdurulamadı: %w", err)
	}

	m.logger.WithField("container_id", containerID).Info("Container durduruldu")
	return nil
}

// Remove removes a container
func (m *Manager) Remove(ctx context.Context, containerID string) error {
	err := m.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("container silinemedi: %w", err)
	}

	m.logger.WithField("container_id", containerID).Info("Container silindi")
	return nil
}

// List lists all containers
func (m *Manager) List(ctx context.Context) ([]*Container, error) {
	containers, err := m.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("container listesi alınamadı: %w", err)
	}

	result := make([]*Container, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		ports := make(map[string]string)
		for _, port := range c.Ports {
			if port.PublicPort > 0 {
				containerPort := strconv.Itoa(int(port.PrivatePort))
				hostPort := strconv.Itoa(int(port.PublicPort))
				ports[containerPort] = hostPort
			}
		}

		result = append(result, &Container{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			Status:  c.Status,
			Ports:   ports,
			Labels:  c.Labels,
			Created: time.Unix(c.Created, 0),
		})
	}

	return result, nil
}

// Get gets a container by ID
func (m *Manager) Get(ctx context.Context, containerID string) (*Container, error) {
	inspect, err := m.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("container bulunamadı: %w", err)
	}

	name := strings.TrimPrefix(inspect.Name, "/")

	ports := make(map[string]string)
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.Ports != nil {
		for port, bindings := range inspect.NetworkSettings.Ports {
			if len(bindings) > 0 {
				containerPort := port.Port()
				hostPort := bindings[0].HostPort
				ports[containerPort] = hostPort
			}
		}
	}

	created, _ := time.Parse(time.RFC3339Nano, inspect.Created)

	var started *time.Time
	if inspect.State.StartedAt != "" {
		if startedTime, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt); err == nil {
			started = &startedTime
		}
	}

	return &Container{
		ID:          inspect.ID,
		Name:        name,
		Image:       inspect.Config.Image,
		Status:      inspect.State.Status,
		Ports:       ports,
		Environment: parseEnvVars(inspect.Config.Env),
		Labels:      inspect.Config.Labels,
		Created:     created,
		Started:     started,
	}, nil
}

// Logs gets container logs with default tail of 100 lines
func (m *Manager) Logs(ctx context.Context, containerID string) (string, error) {
	return m.LogsWithTail(ctx, containerID, 100)
}

// LogsWithTail gets container logs with specified tail count
func (m *Manager) LogsWithTail(ctx context.Context, containerID string, tail int) (string, error) {
	// Limit tail to prevent excessive memory usage
	const maxTail = 10000
	if tail <= 0 {
		tail = 100
	}
	if tail > maxTail {
		tail = maxTail
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := m.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("container logları alınamadı: %w", err)
	}
	defer reader.Close()

	// Use limited buffer to prevent memory issues
	const maxBufferSize = 10 * 1024 * 1024 // 10MB limit
	limitedReader := io.LimitReader(reader, maxBufferSize)
	
	logs, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("loglar okunamadı: %w", err)
	}

	return string(logs), nil
}

// parseEnvVars parses environment variables from Docker format
func parseEnvVars(env []string) map[string]string {
	result := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}
