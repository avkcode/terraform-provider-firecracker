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

// CreateVM sends a request to create a new Firecracker VM
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

// GetVM retrieves information about a VM
func (c *FirecrackerClient) GetVM(ctx context.Context, vmID string) (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/vm/%s", c.BaseURL, vmID)
    tflog.Debug(ctx, "Getting VM info", map[string]interface{}{
        "url": url,
        "id":  vmID,
    })

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request for getting VM info: %w", err)
    }

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request for getting VM info: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, nil // VM not found
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error when getting VM info: status=%d, response=%s", resp.StatusCode, string(body))
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode VM info response: %w", err)
    }

    return result, nil
}

// DeleteVM sends a request to delete a Firecracker VM
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

// UpdateVM sends a request to update a Firecracker VM
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
