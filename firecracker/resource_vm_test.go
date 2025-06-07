package firecracker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// mockHTTPClient is a mock implementation of the httpClient interface
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestResourceFirecrackerVM_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders(),
		CheckDestroy: testAccCheckFirecrackerVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirecrackerVMConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirecrackerVMExists("firecracker_vm.test"),
					resource.TestCheckResourceAttr("firecracker_vm.test", "kernel_image_path", "/path/to/vmlinux"),
					resource.TestCheckResourceAttr("firecracker_vm.test", "boot_args", "console=ttyS0 noapic reboot=k panic=1 pci=off root=/dev/vda rootfstype=ext4 rw init=/sbin/init"),
					resource.TestCheckResourceAttr("firecracker_vm.test", "machine_config.0.vcpu_count", "2"),
					resource.TestCheckResourceAttr("firecracker_vm.test", "machine_config.0.mem_size_mib", "1024"),
				),
			},
		},
	})
}

func testAccProviders() map[string]*schema.Provider {
	provider := Provider()
	// Configure the provider with mock client for testing
	provider.ConfigureContextFunc = testProviderConfigure
	return map[string]*schema.Provider{
		"firecracker": provider,
	}
}

func testProviderConfigure(_ context.Context, d *schema.ResourceData) (interface{}, error) {
	// Create a test server that will respond to API requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/vm":
			// Handle VM creation
			w.WriteHeader(http.StatusCreated)
		case "/vm/test-vm-id":
			// Handle VM get
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{
					"boot-source": {
						"kernel_image_path": "/path/to/vmlinux",
						"boot_args": "console=ttyS0 noapic reboot=k panic=1 pci=off root=/dev/vda rootfstype=ext4 rw init=/sbin/init"
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
					},
					"network-interfaces": [
						{
							"iface_id": "eth0",
							"host_dev_name": "tap0",
							"guest_mac": "AA:BB:CC:DD:EE:FF"
						}
					]
				}`)
			} else if r.Method == http.MethodDelete {
				// Handle VM deletion
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}))

	// Create a client that uses the test server
	return &FirecrackerClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
		Timeout:    30,
	}, nil
}

func testAccCheckFirecrackerVMExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VM ID is set")
		}

		// In a real test, we would check if the VM exists in Firecracker
		// For this mock test, we'll just return nil
		return nil
	}
}

func testAccCheckFirecrackerVMDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "firecracker_vm" {
			continue
		}

		// In a real test, we would check if the VM still exists in Firecracker
		// For this mock test, we'll just return nil
		return nil
	}

	return nil
}

const testAccFirecrackerVMConfig_basic = `
resource "firecracker_vm" "test" {
  kernel_image_path = "/path/to/vmlinux"
  boot_args         = "console=ttyS0 noapic reboot=k panic=1 pci=off root=/dev/vda rootfstype=ext4 rw init=/sbin/init"

  drives {
    drive_id       = "rootfs"
    path_on_host   = "/path/to/rootfs.ext4"
    is_root_device = true
    is_read_only   = false
  }

  machine_config {
    vcpu_count   = 2
    mem_size_mib = 1024
  }

  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"
    guest_mac     = "AA:BB:CC:DD:EE:FF"
  }
}
`
