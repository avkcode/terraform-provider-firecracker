package firecracker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreateVM(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Check if the request is as expected
			if req.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", req.Method)
			}
			if req.URL.String() != "http://localhost:8080/vm" {
				t.Errorf("Expected URL http://localhost:8080/vm, got %s", req.URL.String())
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
			}

			// Return a successful response
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	// Create a client with the mock HTTP client
	client := &FirecrackerClient{
		BaseURL:    "http://localhost:8080",
		HTTPClient: mockClient,
	}

	// Create a VM
	config := map[string]interface{}{
		"boot-source": map[string]interface{}{
			"kernel_image_path": "/path/to/vmlinux",
			"boot_args":         "console=ttyS0 reboot=k panic=1 pci=off",
		},
		"drives": []map[string]interface{}{
			{
				"drive_id":       "rootfs",
				"path_on_host":   "/path/to/rootfs.ext4",
				"is_root_device": true,
				"is_read_only":   false,
			},
		},
		"machine-config": map[string]interface{}{
			"vcpu_count":   2,
			"mem_size_mib": 1024,
		},
		"vm-id": "test-vm",
	}

	err := client.CreateVM(context.Background(), config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetVM(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Check if the request is as expected
			if req.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", req.Method)
			}
			if req.URL.String() != "http://localhost:8080/vm/test-vm" {
				t.Errorf("Expected URL http://localhost:8080/vm/test-vm, got %s", req.URL.String())
			}

			// Return a successful response with VM info
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"boot-source": {
						"kernel_image_path": "/path/to/vmlinux",
						"boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
					},
					"drives": [
						{
							"drive_id": "rootfs",
							"path_on_host": "/path/to/rootfs.ext4",
							"is_root_device": true,
							"is_read_only": false
						}
					],
					"machine-config": {
						"vcpu_count": 2,
						"mem_size_mib": 1024
					}
				}`)),
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			}, nil
		},
	}

	// Create a client with the mock HTTP client
	client := &FirecrackerClient{
		BaseURL:    "http://localhost:8080",
		HTTPClient: mockClient,
	}

	// Get VM info
	vmInfo, err := client.GetVM(context.Background(), "test-vm")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check if the VM info is as expected
	bootSource, ok := vmInfo["boot-source"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected boot-source to be a map, got %T", vmInfo["boot-source"])
	}
	if bootSource["kernel_image_path"] != "/path/to/vmlinux" {
		t.Errorf("Expected kernel_image_path to be /path/to/vmlinux, got %s", bootSource["kernel_image_path"])
	}
}

func TestDeleteVM(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Check if the request is as expected
			if req.Method != http.MethodDelete {
				t.Errorf("Expected DELETE request, got %s", req.Method)
			}
			if req.URL.String() != "http://localhost:8080/vm/test-vm" {
				t.Errorf("Expected URL http://localhost:8080/vm/test-vm, got %s", req.URL.String())
			}

			// Return a successful response
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	// Create a client with the mock HTTP client
	client := &FirecrackerClient{
		BaseURL:    "http://localhost:8080",
		HTTPClient: mockClient,
	}

	// Delete VM
	err := client.DeleteVM(context.Background(), "test-vm")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUpdateVM(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Check if the request is as expected
			if req.Method != http.MethodPatch {
				t.Errorf("Expected PATCH request, got %s", req.Method)
			}
			if req.URL.String() != "http://localhost:8080/vm/test-vm" {
				t.Errorf("Expected URL http://localhost:8080/vm/test-vm, got %s", req.URL.String())
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
			}

			// Check request body
			body, _ := io.ReadAll(req.Body)
			if !strings.Contains(string(body), "machine-config") {
				t.Errorf("Expected request body to contain machine-config, got %s", string(body))
			}

			// Return a successful response
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	// Create a client with the mock HTTP client
	client := &FirecrackerClient{
		BaseURL:    "http://localhost:8080",
		HTTPClient: mockClient,
	}

	// Update VM
	config := map[string]interface{}{
		"machine-config": map[string]interface{}{
			"vcpu_count":   4,
			"mem_size_mib": 2048,
		},
	}

	err := client.UpdateVM(context.Background(), "test-vm", config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
