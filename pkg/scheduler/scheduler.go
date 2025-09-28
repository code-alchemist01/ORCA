package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"orca/pkg/container"

	"github.com/sirupsen/logrus"
)

// Deployment represents a deployment
type Deployment struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	Spec      container.DeploymentSpec  `json:"spec"`
	Status    string                    `json:"status"`
	Replicas  []*container.Container    `json:"replicas"`
	Created   time.Time                 `json:"created"`
}

// Service represents a service
type Service struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Spec      container.ServiceSpec  `json:"spec"`
	Endpoints []string               `json:"endpoints"`
	Status    string                 `json:"status"`
	Created   time.Time              `json:"created"`
}

// Scheduler manages deployments and services
type Scheduler struct {
	containerManager *container.Manager
	deployments      map[string]*Deployment
	services         map[string]*Service
	mutex            sync.RWMutex
	logger           *logrus.Logger
}

// NewScheduler creates a new scheduler
func NewScheduler(containerManager *container.Manager, logger *logrus.Logger) *Scheduler {
	return &Scheduler{
		containerManager: containerManager,
		deployments:      make(map[string]*Deployment),
		services:         make(map[string]*Service),
		logger:           logger,
	}
}

// CreateDeployment creates a new deployment
func (s *Scheduler) CreateDeployment(ctx context.Context, spec container.DeploymentSpec) (*Deployment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if deployment already exists
	for _, d := range s.deployments {
		if d.Name == spec.Name {
			return nil, fmt.Errorf("deployment zaten mevcut: %s", spec.Name)
		}
	}

	deployment := &Deployment{
		ID:       generateID(),
		Name:     spec.Name,
		Spec:     spec,
		Status:   "creating",
		Replicas: make([]*container.Container, 0, spec.Replicas),
		Created:  time.Now(),
	}

	// Create containers for replicas
	for i := 0; i < spec.Replicas; i++ {
		containerSpec := spec.Container
		containerSpec.Name = fmt.Sprintf("%s-%d", spec.Name, i)

		// Assign unique ports for each replica
		if containerSpec.Ports != nil {
			ports := make(map[string]string)
			for containerPort, baseHostPort := range containerSpec.Ports {
				hostPort := fmt.Sprintf("%d", mustParseInt(baseHostPort)+i)
				ports[containerPort] = hostPort
			}
			containerSpec.Ports = ports
		}

		c, err := s.containerManager.Create(ctx, containerSpec)
		if err != nil {
			// Cleanup created containers on error
			s.cleanupDeployment(ctx, deployment)
			return nil, fmt.Errorf("container oluşturulamadı (replica %d): %w", i, err)
		}

		err = s.containerManager.Start(ctx, c.ID)
		if err != nil {
			// Cleanup created containers on error
			s.cleanupDeployment(ctx, deployment)
			return nil, fmt.Errorf("container başlatılamadı (replica %d): %w", i, err)
		}

		c.Status = "running"
		deployment.Replicas = append(deployment.Replicas, c)
	}

	deployment.Status = "running"
	s.deployments[deployment.ID] = deployment

	s.logger.WithFields(logrus.Fields{
		"deployment_id": deployment.ID,
		"name":          deployment.Name,
		"replicas":      spec.Replicas,
	}).Info("Deployment oluşturuldu")

	return deployment, nil
}

// GetDeployment gets a deployment by name
func (s *Scheduler) GetDeployment(name string) (*Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, d := range s.deployments {
		if d.Name == name {
			return d, nil
		}
	}

	return nil, fmt.Errorf("deployment bulunamadı: %s", name)
}

// ListDeployments lists all deployments
func (s *Scheduler) ListDeployments() []*Deployment {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	deployments := make([]*Deployment, 0, len(s.deployments))
	for _, d := range s.deployments {
		deployments = append(deployments, d)
	}

	return deployments
}

// DeleteDeployment deletes a deployment
func (s *Scheduler) DeleteDeployment(ctx context.Context, name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var deployment *Deployment
	var deploymentID string
	
	for id, d := range s.deployments {
		if d.Name == name {
			deployment = d
			deploymentID = id
			break
		}
	}

	if deployment == nil {
		return fmt.Errorf("deployment bulunamadı: %s", name)
	}

	// Stop and remove all containers
	err := s.cleanupDeployment(ctx, deployment)
	if err != nil {
		return fmt.Errorf("deployment temizlenemedi: %w", err)
	}

	delete(s.deployments, deploymentID)

	s.logger.WithFields(logrus.Fields{
		"deployment_id": deploymentID,
		"name":          name,
	}).Info("Deployment silindi")

	return nil
}

// CreateService creates a new service
func (s *Scheduler) CreateService(ctx context.Context, spec container.ServiceSpec) (*Service, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if service already exists
	for _, svc := range s.services {
		if svc.Name == spec.Name {
			return nil, fmt.Errorf("service zaten mevcut: %s", spec.Name)
		}
	}

	service := &Service{
		ID:        generateID(),
		Name:      spec.Name,
		Spec:      spec,
		Endpoints: []string{},
		Status:    "active",
		Created:   time.Now(),
	}

	s.services[service.ID] = service

	s.logger.WithFields(logrus.Fields{
		"service_id": service.ID,
		"name":       service.Name,
		"type":       spec.Type,
	}).Info("Service oluşturuldu")

	return service, nil
}

// GetService gets a service by name
func (s *Scheduler) GetService(name string) (*Service, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, svc := range s.services {
		if svc.Name == name {
			return svc, nil
		}
	}

	return nil, fmt.Errorf("service bulunamadı: %s", name)
}

// ListServices lists all services
func (s *Scheduler) ListServices() []*Service {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	services := make([]*Service, 0, len(s.services))
	for _, svc := range s.services {
		services = append(services, svc)
	}

	return services
}

// DeleteService deletes a service
func (s *Scheduler) DeleteService(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var serviceID string
	for id, svc := range s.services {
		if svc.Name == name {
			serviceID = id
			break
		}
	}

	if serviceID == "" {
		return fmt.Errorf("service bulunamadı: %s", name)
	}

	delete(s.services, serviceID)

	s.logger.WithFields(logrus.Fields{
		"service_id": serviceID,
		"name":       name,
	}).Info("Service silindi")

	return nil
}

// cleanupDeployment removes all containers in a deployment
func (s *Scheduler) cleanupDeployment(ctx context.Context, deployment *Deployment) error {
	for _, c := range deployment.Replicas {
		if err := s.containerManager.Stop(ctx, c.ID); err != nil {
			s.logger.WithError(err).WithField("container_id", c.ID).Warn("Container durdurulamadı")
		}
		if err := s.containerManager.Remove(ctx, c.ID); err != nil {
			s.logger.WithError(err).WithField("container_id", c.ID).Warn("Container silinemedi")
		}
	}
	return nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// mustParseInt parses string to int, panics on error
func mustParseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}