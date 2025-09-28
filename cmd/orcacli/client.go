package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"orca/pkg/container"
	"orca/pkg/scheduler"
)

// HTTP client functions

func createContainer(spec container.ContainerSpec) (*container.Container, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/containers", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var c container.Container
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func listContainers() ([]*container.Container, error) {
	resp, err := http.Get(serverURL + "/containers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var containers []*container.Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}

	return containers, nil
}

func startContainer(containerID string) error {
	resp, err := http.Post(serverURL+"/containers/"+containerID+"/start", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func stopContainer(containerID string) error {
	resp, err := http.Post(serverURL+"/containers/"+containerID+"/stop", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func removeContainer(containerID string) error {
	req, err := http.NewRequest("DELETE", serverURL+"/containers/"+containerID+"/remove", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func inspectContainer(containerID string) (*container.Container, error) {
	resp, err := http.Get(serverURL + "/containers/" + containerID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var c container.Container
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func getContainerLogs(containerID string, tail int) (string, error) {
	url := fmt.Sprintf("%s/containers/%s/logs?tail=%d", serverURL, containerID, tail)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func createDeployment(spec container.DeploymentSpec) (*scheduler.Deployment, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/deployments", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var deployment scheduler.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

func listDeployments() ([]*scheduler.Deployment, error) {
	resp, err := http.Get(serverURL + "/deployments")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var deployments []*scheduler.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, err
	}

	return deployments, nil
}

func deleteDeployment(name string) error {
	req, err := http.NewRequest("DELETE", serverURL+"/deployments/"+name, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func createService(spec container.ServiceSpec) (*scheduler.Service, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/services", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var service scheduler.Service
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, err
	}

	return &service, nil
}

func listServices() ([]*scheduler.Service, error) {
	resp, err := http.Get(serverURL + "/services")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var services []*scheduler.Service
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, err
	}

	return services, nil
}

func deleteService(name string) error {
	req, err := http.NewRequest("DELETE", serverURL+"/services/"+name, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func getStats() (map[string]interface{}, error) {
	resp, err := http.Get(serverURL + "/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Utility functions for formatting output

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}

func formatPorts(ports map[string]string) string {
	if len(ports) == 0 {
		return "-"
	}

	var portStrings []string
	for containerPort, hostPort := range ports {
		portStrings = append(portStrings, fmt.Sprintf("%s:%s", hostPort, containerPort))
	}

	return strings.Join(portStrings, ", ")
}

func formatServicePorts(ports []container.ServicePort) string {
	if len(ports) == 0 {
		return "-"
	}

	var portStrings []string
	for _, port := range ports {
		if port.TargetPort != 0 {
			portStrings = append(portStrings, fmt.Sprintf("%d:%d", port.Port, port.TargetPort))
		} else {
			portStrings = append(portStrings, strconv.Itoa(port.Port))
		}
	}

	return strings.Join(portStrings, ", ")
}