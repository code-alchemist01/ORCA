package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"orca/pkg/scheduler"

	"github.com/sirupsen/logrus"
)

// Storage handles persistent storage of deployments and services
type Storage struct {
	dataDir string
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string, logger *logrus.Logger) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("data dizini oluşturulamadı: %w", err)
	}

	// Create subdirectories
	dirs := []string{"deployments", "services", "containers"}
	for _, dir := range dirs {
		dirPath := filepath.Join(dataDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("alt dizin oluşturulamadı (%s): %w", dir, err)
		}
	}

	return &Storage{
		dataDir: dataDir,
		logger:  logger,
	}, nil
}

// SaveDeployment saves a deployment to storage
func (s *Storage) SaveDeployment(deployment *scheduler.Deployment) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.MarshalIndent(deployment, "", "  ")
	if err != nil {
		return fmt.Errorf("deployment serialize edilemedi: %w", err)
	}

	filePath := filepath.Join(s.dataDir, "deployments", fmt.Sprintf("%s.json", deployment.ID))
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("deployment kaydedilemedi: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment_id": deployment.ID,
		"name":          deployment.Name,
		"file":          filePath,
	}).Debug("Deployment kaydedildi")

	return nil
}

// LoadDeployment loads a deployment from storage
func (s *Storage) LoadDeployment(id string) (*scheduler.Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filePath := filepath.Join(s.dataDir, "deployments", fmt.Sprintf("%s.json", id))
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("deployment bulunamadı: %s", id)
		}
		return nil, fmt.Errorf("deployment okunamadı: %w", err)
	}

	var deployment scheduler.Deployment
	if err := json.Unmarshal(data, &deployment); err != nil {
		return nil, fmt.Errorf("deployment parse edilemedi: %w", err)
	}

	return &deployment, nil
}

// LoadAllDeployments loads all deployments from storage
func (s *Storage) LoadAllDeployments() ([]*scheduler.Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	deploymentsDir := filepath.Join(s.dataDir, "deployments")
	files, err := ioutil.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("deployments dizini okunamadı: %w", err)
	}

	var deployments []*scheduler.Deployment
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(deploymentsDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Deployment dosyası okunamadı")
			continue
		}

		var deployment scheduler.Deployment
		if err := json.Unmarshal(data, &deployment); err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Deployment parse edilemedi")
			continue
		}

		deployments = append(deployments, &deployment)
	}

	return deployments, nil
}

// DeleteDeployment deletes a deployment from storage
func (s *Storage) DeleteDeployment(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filePath := filepath.Join(s.dataDir, "deployments", fmt.Sprintf("%s.json", id))
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("deployment bulunamadı: %s", id)
		}
		return fmt.Errorf("deployment silinemedi: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment_id": id,
		"file":          filePath,
	}).Debug("Deployment silindi")

	return nil
}

// SaveService saves a service to storage
func (s *Storage) SaveService(service *scheduler.Service) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.MarshalIndent(service, "", "  ")
	if err != nil {
		return fmt.Errorf("service serialize edilemedi: %w", err)
	}

	filePath := filepath.Join(s.dataDir, "services", fmt.Sprintf("%s.json", service.ID))
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("service kaydedilemedi: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"service_id": service.ID,
		"name":       service.Name,
		"file":       filePath,
	}).Debug("Service kaydedildi")

	return nil
}

// LoadService loads a service from storage
func (s *Storage) LoadService(id string) (*scheduler.Service, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filePath := filepath.Join(s.dataDir, "services", fmt.Sprintf("%s.json", id))
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("service bulunamadı: %s", id)
		}
		return nil, fmt.Errorf("service okunamadı: %w", err)
	}

	var service scheduler.Service
	if err := json.Unmarshal(data, &service); err != nil {
		return nil, fmt.Errorf("service parse edilemedi: %w", err)
	}

	return &service, nil
}

// LoadAllServices loads all services from storage
func (s *Storage) LoadAllServices() ([]*scheduler.Service, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	servicesDir := filepath.Join(s.dataDir, "services")
	files, err := ioutil.ReadDir(servicesDir)
	if err != nil {
		return nil, fmt.Errorf("services dizini okunamadı: %w", err)
	}

	var services []*scheduler.Service
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(servicesDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Service dosyası okunamadı")
			continue
		}

		var service scheduler.Service
		if err := json.Unmarshal(data, &service); err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Service parse edilemedi")
			continue
		}

		services = append(services, &service)
	}

	return services, nil
}

// DeleteService deletes a service from storage
func (s *Storage) DeleteService(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filePath := filepath.Join(s.dataDir, "services", fmt.Sprintf("%s.json", id))
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("service bulunamadı: %s", id)
		}
		return fmt.Errorf("service silinemedi: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"service_id": id,
		"file":       filePath,
	}).Debug("Service silindi")

	return nil
}

// GetStats returns storage statistics
func (s *Storage) GetStats() (map[string]int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := make(map[string]int)

	// Count deployments
	deploymentsDir := filepath.Join(s.dataDir, "deployments")
	deploymentFiles, err := ioutil.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("deployments dizini okunamadı: %w", err)
	}
	
	deploymentCount := 0
	for _, file := range deploymentFiles {
		if filepath.Ext(file.Name()) == ".json" {
			deploymentCount++
		}
	}
	stats["deployments"] = deploymentCount

	// Count services
	servicesDir := filepath.Join(s.dataDir, "services")
	serviceFiles, err := ioutil.ReadDir(servicesDir)
	if err != nil {
		return nil, fmt.Errorf("services dizini okunamadı: %w", err)
	}
	
	serviceCount := 0
	for _, file := range serviceFiles {
		if filepath.Ext(file.Name()) == ".json" {
			serviceCount++
		}
	}
	stats["services"] = serviceCount

	return stats, nil
}