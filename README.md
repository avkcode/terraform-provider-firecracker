# Terraform Provider for Firecracker

This Terraform provider allows you to manage [Firecracker](https://firecracker-microvm.github.io/) microVMs through Terraform. Firecracker is a virtualization technology that enables you to implement lightweight virtual machines with minimal overhead.

## Documentation

Full documentation is available in the [docs](./docs) directory, including:
- [Installation Guide](docs/guides/installation.md)
- [Getting Started Guide](docs/guides/getting-started.md)
- [Firecracker Setup Guide](docs/guides/firecracker-setup.md)
- [Troubleshooting Guide](docs/guides/troubleshooting.md)
- [Resource Documentation](docs/resources/vm.md)
- [Data Source Documentation](docs/data-sources/vm.md)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (to build the provider plugin)
- [Firecracker](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md) running with API access
- [socat](http://www.dest-unreach.org/socat/) for forwarding HTTP traffic to the Firecracker Unix socket
- Linux with KVM support (Firecracker only runs on Linux)

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/terraform-provider-firecracker/terraform-provider-firecracker.git
cd terraform-provider-firecracker

# Build and install the provider
make build

# Set up the Firecracker environment
make setup
```

### Configuration

```hcl
terraform {
  required_providers {
    firecracker = {
      source  = "terraform-provider-firecracker/firecracker"
      version = "0.1.0"
    }
  }
}

# Configure the Firecracker provider
provider "firecracker" {
  base_url = "http://localhost:8080"  # URL to your Firecracker API
  timeout  = 30                       # Optional: timeout in seconds (default: 30)
}

# Define a Firecracker VM
resource "firecracker_vm" "example" {
  kernel_image_path = "/path/to/vmlinux"
  boot_args         = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"

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
    guest_mac     = "AA:BB:CC:DD:EE:FF"  # Optional
  }
}

# Query an existing VM
data "firecracker_vm" "existing" {
  vm_id = firecracker_vm.example.id
}
```

## Development

### Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22
- [Firecracker](https://github.com/firecracker-microvm/firecracker/releases) binary
- [socat](http://www.dest-unreach.org/socat/) for API forwarding

### Makefile Utilities

The project includes a comprehensive Makefile with useful commands:

```bash
# Check all dependencies
make verify

# Build the provider
make build

# Run a test configuration
make run

# Check environment status
make status

# Set up the environment
make setup

# Tear down the environment
make teardown

# Get help on all commands
make help
```

### Testing

```bash
# Run basic tests
make test

# Run a more complex test with remote-exec
make test-remote-exec
```

## Troubleshooting

If you encounter issues, check the [Troubleshooting Guide](docs/guides/troubleshooting.md) or run:

```bash
# Check the status of all components
make status

# Restart the environment
make teardown && make setup
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
