# Terraform Provider for Firecracker

This Terraform provider allows you to manage [Firecracker](https://firecracker-microvm.github.io/) microVMs through Terraform.

## Documentation

Full documentation is available in the [docs](./docs) directory.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (to build the provider plugin)
- [Firecracker](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md) running with API access

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/avkcode/terraform-provider-firecracker.git
cd terraform-provider-firecracker

# Build and install the provider
make build
```

### Configuration

```hcl
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
```

## Development

### Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

### Building

```bash
make build
```

### Testing

```bash
make test
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
