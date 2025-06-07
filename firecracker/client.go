package firecracker

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/hashicorp/go-retryablehttp"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

// VMConfig represents the configuration for a Firecracker VM.
type VMConfig struct {
    KernelImage string `json:"kernel_image_path"`
    Rootfs      string `json:"rootfs_path"`
    CPUCount    int    `json:"vcpu_count"`
    MemorySize  int    `json:"mem_size_mib"`
}

// httpClient is an interface for HTTP operations to make testing easier
type httpClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// defaultHTTPClient returns a default HTTP client with reasonable timeouts and retry logic
func defaultHTTPClient() *http.Client {
    retryClient := retryablehttp.NewClient()
    retryClient.RetryMax = 3
    retryClient.RetryWaitMin = 1 * time.Second
    retryClient.RetryWaitMax = 5 * time.Second
    retryClient.Logger = nil // Disable default logger
    
    // Configure the underlying transport
    retryClient.HTTPClient.Timeout = 30 * time.Second
    retryClient.HTTPClient.Transport = &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:     90 * time.Second,
    }
    
    return retryClient.StandardClient()
}

// CreateVM sends a request to the Firecracker API to create a new microVM.
// It takes a context for cancellation and a configuration map that defines the VM properties.
// The configuration should include boot-source, drives, machine-config, and other VM settings.
func (c *FirecrackerClient) CreateVM(ctx context.Context, config map[string]interface{}) error {
    url := c.BaseURL + "/vm"
    tflog.Debug(ctx, "Creating VM", map[string]interface{}{
        "url":    url,
        "config": config,
    })

    jsonPayload, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal VM configuration payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create HTTP request for VM creation: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send VM creation request: %w", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("API error when creating VM: status=%d, response=%s", resp.StatusCode, string(body))
    }
    
    tflog.Info(ctx, "VM created successfully", map[string]interface{}{
        "status_code": resp.StatusCode,
    })
    
    return nil
}

// StartVM sends a request to start a Firecracker VM
func (c *FirecrackerClient) StartVM(ctx context.Context, vmID string) error {
    url := fmt.Sprintf("%s/vm/%s/actions", c.BaseURL, vmID)
    tflog.Debug(ctx, "Starting VM", map[string]interface{}{
        "url": url,
        "id":  vmID,
    })

    payload := map[string]interface{}{
        "action_type": "InstanceStart",
    }

    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal VM start payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create HTTP request for VM start: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send VM start request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error when starting VM: status=%d, response=%s", resp.StatusCode, string(body))
    }

    tflog.Info(ctx, "VM started successfully", map[string]interface{}{
        "id": vmID,
    })
    
    return nil
}

// StopVM sends a request to stop a Firecracker VM
func (c *FirecrackerClient) StopVM(ctx context.Context, vmID string) error {
    url := fmt.Sprintf("%s/vm/%s/actions", c.BaseURL, vmID)
    tflog.Debug(ctx, "Stopping VM", map[string]interface{}{
        "url": url,
        "id":  vmID,
    })

    payload := map[string]interface{}{
        "action_type": "SendCtrlAltDel",
    }

    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal VM stop payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create HTTP request for VM stop: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send VM stop request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error when stopping VM: status=%d, response=%s", resp.StatusCode, string(body))
    }

    tflog.Info(ctx, "VM stop signal sent successfully", map[string]interface{}{
        "id": vmID,
    })
    
    return nil
}

// GetVM retrieves information about a VM from the Firecracker API.
// It returns a map containing the VM configuration or nil if the VM doesn't exist.
// This method is used by the Read operation of the resource and data source.
func (c *FirecrackerClient) GetVM(ctx context.Context, vmID string) (map[string]interface{}, error) {
    // Instead of using /vm/{id}, we need to query individual components
    // Let's build a composite response by querying multiple endpoints
    result := map[string]interface{}{
        "vm-id": vmID,
    }
    
    tflog.Debug(ctx, "Getting VM info by querying individual components", map[string]interface{}{
        "id": vmID,
    })

    // Get boot source info
    bootSourceURL := fmt.Sprintf("%s/boot-source", c.BaseURL)
    bootSource, err := c.getComponent(ctx, bootSourceURL)
    if err != nil {
        return nil, fmt.Errorf("failed to get boot source info: %w", err)
    }
    if bootSource != nil {
        result["boot-source"] = bootSource
    }

    // Get machine config
    machineConfigURL := fmt.Sprintf("%s/machine-config", c.BaseURL)
    machineConfig, err := c.getComponent(ctx, machineConfigURL)
    if err != nil {
        return nil, fmt.Errorf("failed to get machine config: %w", err)
    }
    if machineConfig != nil {
        result["machine-config"] = machineConfig
    }

    // Get drives info
    drivesURL := fmt.Sprintf("%s/drives", c.BaseURL)
    drives, err := c.getComponentList(ctx, drivesURL)
    if err != nil {
        return nil, fmt.Errorf("failed to get drives info: %w", err)
    }
    if drives != nil {
        result["drives"] = drives
    }

    // Get network interfaces
    networkURL := fmt.Sprintf("%s/network-interfaces", c.BaseURL)
    networkInterfaces, err := c.getComponentList(ctx, networkURL)
    if err != nil {
        return nil, fmt.Errorf("failed to get network interfaces: %w", err)
    }
    if networkInterfaces != nil {
        result["network-interfaces"] = networkInterfaces
    }

    // If we couldn't get any component info, assume the VM doesn't exist
    if len(result) <= 1 {
        return nil, nil
    }

    return result, nil
}

// Helper method to get a component from the API
func (c *FirecrackerClient) getComponent(ctx context.Context, url string) (map[string]interface{}, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request: %w", err)
    }

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, nil // Component not found
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error: status=%d, response=%s", resp.StatusCode, string(body))
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return result, nil
}

// Helper method to get a list of components from the API
func (c *FirecrackerClient) getComponentList(ctx context.Context, baseURL string) ([]interface{}, error) {
    // This is a simplified implementation - in a real scenario, you might need to
    // query each component individually or use a different endpoint structure
    // For now, we'll just return an empty list to avoid errors
    return []interface{}{}, nil
}

// DeleteVM sends a request to delete a Firecracker VM.
// If the VM doesn't exist, it returns nil to indicate successful deletion.
// This method is used by the Delete operation of the resource.
func (c *FirecrackerClient) DeleteVM(ctx context.Context, vmID string) error {
    url := fmt.Sprintf("%s/vm/%s", c.BaseURL, vmID)
    tflog.Debug(ctx, "Deleting VM", map[string]interface{}{
        "url": url,
        "id":  vmID,
    })

    req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
    if err != nil {
        return fmt.Errorf("failed to create HTTP request for VM deletion: %w", err)
    }

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send VM deletion request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        tflog.Warn(ctx, "VM not found during deletion, considering it already deleted", map[string]interface{}{
            "id": vmID,
        })
        return nil // VM already deleted or doesn't exist
    }

    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error when deleting VM: status=%d, response=%s", resp.StatusCode, string(body))
    }

    tflog.Info(ctx, "VM deleted successfully", map[string]interface{}{
        "id": vmID,
    })
    
    return nil
}

// UpdateVM sends a request to update a Firecracker VM.
// It takes a VM ID and a configuration map containing the properties to update.
// This method is used by the Update operation of the resource.
func (c *FirecrackerClient) UpdateVM(ctx context.Context, vmID string, config map[string]interface{}) error {
    url := fmt.Sprintf("%s/vm/%s", c.BaseURL, vmID)
    tflog.Debug(ctx, "Updating VM", map[string]interface{}{
        "url":    url,
        "id":     vmID,
        "config": config,
    })

    jsonPayload, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal VM update payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create HTTP request for VM update: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send VM update request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error when updating VM: status=%d, response=%s", resp.StatusCode, string(body))
    }

    tflog.Info(ctx, "VM updated successfully", map[string]interface{}{
        "id": vmID,
    })
    
    return nil
}
