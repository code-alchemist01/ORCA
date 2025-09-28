package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"orca/pkg/container"

	"github.com/gorilla/mux"
)

// healthHandler handles health check requests
func (s *OrcaServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
		"service": "orca-orchestrator",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// listContainersHandler handles listing containers
func (s *OrcaServer) listContainersHandler(w http.ResponseWriter, r *http.Request) {
	containers, err := s.containerManager.List(r.Context())
	if err != nil {
		s.logger.WithError(err).Error("Container listesi alınamadı")
		http.Error(w, "Container listesi alınamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// createContainerHandler handles container creation
func (s *OrcaServer) createContainerHandler(w http.ResponseWriter, r *http.Request) {
	var spec container.ContainerSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
		return
	}

	// Input validation
	if spec.Name == "" {
		http.Error(w, "Container adı boş olamaz", http.StatusBadRequest)
		return
	}

	if spec.Image == "" {
		http.Error(w, "Container image boş olamaz", http.StatusBadRequest)
		return
	}

	// Validate ports if specified
	for hostPortStr, containerPortStr := range spec.Ports {
		// Parse host port
		hostPort := 0
		if hostPortStr != "" {
			if hp, err := strconv.Atoi(hostPortStr); err != nil {
				http.Error(w, "Geçersiz host port formatı", http.StatusBadRequest)
				return
			} else {
				hostPort = hp
			}
		}

		// Parse container port
		containerPort := 0
		if containerPortStr != "" {
			if cp, err := strconv.Atoi(containerPortStr); err != nil {
				http.Error(w, "Geçersiz container port formatı", http.StatusBadRequest)
				return
			} else {
				containerPort = cp
			}
		}

		// Validate port ranges
		if hostPort < 1 || hostPort > 65535 {
			http.Error(w, "Host port numarası 1-65535 arasında olmalıdır", http.StatusBadRequest)
			return
		}
		if containerPort < 1 || containerPort > 65535 {
			http.Error(w, "Container port numarası 1-65535 arasında olmalıdır", http.StatusBadRequest)
			return
		}
	}

	c, err := s.containerManager.Create(r.Context(), spec)
	if err != nil {
		s.logger.WithError(err).Error("Container oluşturulamadı")
		http.Error(w, "Container oluşturulamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// getContainerHandler handles getting a specific container
func (s *OrcaServer) getContainerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Resolve name to container ID
	containerID, err := s.resolveContainerID(r.Context(), name)
	if err != nil {
		s.logger.WithError(err).Error("Container bulunamadı")
		http.Error(w, "Container bulunamadı", http.StatusNotFound)
		return
	}

	c, err := s.containerManager.Get(r.Context(), containerID)
	if err != nil {
		s.logger.WithError(err).Error("Container bulunamadı")
		http.Error(w, "Container bulunamadı", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// resolveContainerID resolves container name to ID
func (s *OrcaServer) resolveContainerID(ctx context.Context, nameOrID string) (string, error) {
	// First try to get container by name/ID directly
	container, err := s.containerManager.Get(ctx, nameOrID)
	if err == nil {
		return container.ID, nil
	}

	// If that fails, try to find by name in the container list
	containers, err := s.containerManager.List(ctx)
	if err != nil {
		return "", err
	}

	for _, c := range containers {
		if c.Name == nameOrID || c.ID == nameOrID || strings.HasPrefix(c.ID, nameOrID) {
			return c.ID, nil
		}
	}

	return "", fmt.Errorf("container bulunamadı: %s", nameOrID)
}

// startContainerHandler handles starting a container
func (s *OrcaServer) startContainerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Resolve name to container ID
	containerID, err := s.resolveContainerID(r.Context(), name)
	if err != nil {
		s.logger.WithError(err).Error("Container bulunamadı")
		http.Error(w, "Container bulunamadı", http.StatusNotFound)
		return
	}

	if err := s.containerManager.Start(r.Context(), containerID); err != nil {
		s.logger.WithError(err).Error("Container başlatılamadı")
		http.Error(w, "Container başlatılamadı", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// stopContainerHandler handles stopping a container
func (s *OrcaServer) stopContainerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Resolve name to container ID
	containerID, err := s.resolveContainerID(r.Context(), name)
	if err != nil {
		s.logger.WithError(err).Error("Container bulunamadı")
		http.Error(w, "Container bulunamadı", http.StatusNotFound)
		return
	}

	if err := s.containerManager.Stop(r.Context(), containerID); err != nil {
		s.logger.WithError(err).Error("Container durdurulamadı")
		http.Error(w, "Container durdurulamadı", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// removeContainerHandler handles removing a container
func (s *OrcaServer) removeContainerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Resolve name to container ID
	containerID, err := s.resolveContainerID(r.Context(), name)
	if err != nil {
		s.logger.WithError(err).Error("Container bulunamadı")
		http.Error(w, "Container bulunamadı", http.StatusNotFound)
		return
	}

	if err := s.containerManager.Remove(r.Context(), containerID); err != nil {
		s.logger.WithError(err).Error("Container silinemedi")
		http.Error(w, "Container silinemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

// containerLogsHandler handles getting container logs
func (s *OrcaServer) containerLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Resolve name to container ID
	containerID, err := s.resolveContainerID(r.Context(), name)
	if err != nil {
		s.logger.WithError(err).Error("Container bulunamadı")
		http.Error(w, "Container bulunamadı", http.StatusNotFound)
		return
	}

	// Parse tail parameter from query string
	tailStr := r.URL.Query().Get("tail")
	tail := 100 // default value
	if tailStr != "" {
		if parsedTail, err := strconv.Atoi(tailStr); err == nil && parsedTail > 0 {
			tail = parsedTail
		}
	}

	logs, err := s.containerManager.LogsWithTail(r.Context(), containerID, tail)
	if err != nil {
		s.logger.WithError(err).Error("Container logları alınamadı")
		http.Error(w, "Container logları alınamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logs))
}

// listDeploymentsHandler handles listing deployments
func (s *OrcaServer) listDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	deployments := s.scheduler.ListDeployments()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

// createDeploymentHandler handles deployment creation
func (s *OrcaServer) createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	var spec container.DeploymentSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
		return
	}

	// Input validation
	if spec.Name == "" {
		http.Error(w, "Deployment adı boş olamaz", http.StatusBadRequest)
		return
	}

	if spec.Replicas < 1 {
		http.Error(w, "Replica sayısı en az 1 olmalıdır", http.StatusBadRequest)
		return
	}

	if spec.Replicas > 100 {
		http.Error(w, "Replica sayısı en fazla 100 olabilir", http.StatusBadRequest)
		return
	}

	if spec.Container.Name == "" {
		http.Error(w, "Container adı boş olamaz", http.StatusBadRequest)
		return
	}

	if spec.Container.Image == "" {
		http.Error(w, "Container image boş olamaz", http.StatusBadRequest)
		return
	}

	deployment, err := s.scheduler.CreateDeployment(r.Context(), spec)
	if err != nil {
		s.logger.WithError(err).Error("Deployment oluşturulamadı")
		http.Error(w, "Deployment oluşturulamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

// getDeploymentHandler handles getting a specific deployment
func (s *OrcaServer) getDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	deployment, err := s.scheduler.GetDeployment(name)
	if err != nil {
		s.logger.WithError(err).Error("Deployment bulunamadı")
		http.Error(w, "Deployment bulunamadı", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

// deleteDeploymentHandler handles deployment deletion
func (s *OrcaServer) deleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	deployment, err := s.scheduler.GetDeployment(name)
	if err != nil {
		s.logger.WithError(err).Error("Deployment bulunamadı")
		http.Error(w, "Deployment bulunamadı", http.StatusNotFound)
		return
	}

	if err := s.scheduler.DeleteDeployment(r.Context(), name); err != nil {
		s.logger.WithError(err).Error("Deployment silinemedi")
		http.Error(w, "Deployment silinemedi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

// listServicesHandler handles listing services
func (s *OrcaServer) listServicesHandler(w http.ResponseWriter, r *http.Request) {
	services := s.scheduler.ListServices()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// createServiceHandler handles service creation
func (s *OrcaServer) createServiceHandler(w http.ResponseWriter, r *http.Request) {
	var spec container.ServiceSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
		return
	}

	// Input validation
	if spec.Name == "" {
		http.Error(w, "Service adı boş olamaz", http.StatusBadRequest)
		return
	}

	if spec.Type == "" {
		http.Error(w, "Service tipi belirtilmelidir", http.StatusBadRequest)
		return
	}

	if spec.Type != "ClusterIP" && spec.Type != "NodePort" && spec.Type != "LoadBalancer" {
		http.Error(w, "Geçersiz service tipi. Desteklenen tipler: ClusterIP, NodePort, LoadBalancer", http.StatusBadRequest)
		return
	}

	if len(spec.Ports) == 0 {
		http.Error(w, "En az bir port tanımlanmalıdır", http.StatusBadRequest)
		return
	}

	// Validate ports
	for _, port := range spec.Ports {
		if port.Port < 1 || port.Port > 65535 {
			http.Error(w, "Port numarası 1-65535 arasında olmalıdır", http.StatusBadRequest)
			return
		}
		if port.TargetPort < 1 || port.TargetPort > 65535 {
			http.Error(w, "Hedef port numarası 1-65535 arasında olmalıdır", http.StatusBadRequest)
			return
		}
	}

	service, err := s.scheduler.CreateService(r.Context(), spec)
	if err != nil {
		s.logger.WithError(err).Error("Service oluşturulamadı")
		http.Error(w, "Service oluşturulamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// getServiceHandler handles getting a specific service
func (s *OrcaServer) getServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	service, err := s.scheduler.GetService(name)
	if err != nil {
		s.logger.WithError(err).Error("Service bulunamadı")
		http.Error(w, "Service bulunamadı", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// deleteServiceHandler handles service deletion
func (s *OrcaServer) deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	service, err := s.scheduler.GetService(name)
	if err != nil {
		s.logger.WithError(err).Error("Service bulunamadı")
		http.Error(w, "Service bulunamadı", http.StatusNotFound)
		return
	}

	if err := s.scheduler.DeleteService(name); err != nil {
		s.logger.WithError(err).Error("Service silinemedi")
		http.Error(w, "Service silinemedi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// statsHandler handles getting system statistics
func (s *OrcaServer) statsHandler(w http.ResponseWriter, r *http.Request) {
	containers, err := s.containerManager.List(r.Context())
	if err != nil {
		s.logger.WithError(err).Error("Container istatistikleri alınamadı")
		http.Error(w, "İstatistikler alınamadı", http.StatusInternalServerError)
		return
	}

	deployments := s.scheduler.ListDeployments()
	services := s.scheduler.ListServices()

	stats := map[string]interface{}{
		"containers":  len(containers),
		"deployments": len(deployments),
		"services":    len(services),
		"uptime":      "N/A", // Bu daha sonra implement edilebilir
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}