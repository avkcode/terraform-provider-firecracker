package firecracker

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type VMConfig struct {
    KernelImage string `json:"kernel_image_path"`
    Rootfs      string `json:"rootfs_path"`
    CPUCount    int    `json:"vcpu_count"`
    MemorySize  int    `json:"mem_size_mib"`
}

func (c *FirecrackerClient) CreateVM(config map[string]interface{}) error {
    url := c.BaseURL + "/vm"

    jsonPayload, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    return nil
}
