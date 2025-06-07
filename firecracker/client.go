package firecracker

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

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

// defaultHTTPClient returns a default HTTP client with reasonable timeouts
func defaultHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 20,
            IdleConnTimeout:     90 * time.Second,
        },
    }
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
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
    }
    
    tflog.Info(ctx, "VM created successfully", map[string]interface{}{
        "status_code": resp.StatusCode,
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
        return nil, fmt.Errorf("failed to create request: %w", err)
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
        return nil, nil // VM not found
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
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
        return fmt.Errorf("failed to create request: %w", err)
    }

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
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
        return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
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
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := c.HTTPClient
    if client == nil {
        client = defaultHTTPClient()
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
    }

    tflog.Info(ctx, "VM updated successfully", map[string]interface{}{
        "id": vmID,
    })
    
    return nil
}
