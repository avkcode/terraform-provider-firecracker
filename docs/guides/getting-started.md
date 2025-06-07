# Getting Started with the Firecracker Terraform Provider

This guide will help you get started with the Firecracker Terraform Provider by walking through the process of creating, managing, and destroying Firecracker microVMs.

## Prerequisites

Before you begin, ensure you have the following:

1. [Terraform](https://www.terraform.io/downloads.html) >= 1.0
2. [Firecracker](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md) installed
3. A Linux kernel binary (vmlinux format)
4. A root filesystem image (ext4 format)
5. The Firecracker Terraform Provider [installed](installation.md)

## Step 1: Set Up the Environment

The provider includes a Makefile that simplifies environment setup:

```bash
# Check if all dependencies are installed
make verify

# Set up the Firecracker environment (starts Firecracker and socat)
make setup

# Verify the environment is running correctly
make status
```

If you prefer to set up manually:

```bash
# Create a socket for Firecracker
rm -f /tmp/firecracker.sock

# Start Firecracker
firecracker --api-sock /tmp/firecracker.sock &

# Use socat to forward TCP traffic to the Unix socket
socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock &
```

## Step 2: Create Your Terraform Configuration

Create a new directory for your Terraform configuration:

```bash
mkdir my-firecracker-project
cd my-firecracker-project
```

Create a file named `main.tf` with the following content:

```hcl
terraform {
  required_providers {
    firecracker = {
      source = "terraform-provider-firecracker/firecracker"
      version = "0.1.0"
    }
  }
}

provider "firecracker" {
  base_url = "http://localhost:8080"
}

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
  }
}
```

Replace `/path/to/vmlinux` and `/path/to/rootfs.ext4` with the actual paths to your kernel and root filesystem images.

### Understanding the Configuration

Let's break down the configuration:

1. **Provider Block**: Configures the Firecracker provider with the API endpoint.

2. **VM Resource**: Defines a Firecracker microVM with:
   - **kernel_image_path**: Path to the Linux kernel binary
   - **boot_args**: Kernel boot parameters
   - **drives**: Storage configuration (at least one drive with the root filesystem)
   - **machine_config**: VM hardware specifications (CPU and memory)
   - **network_interfaces**: Network configuration (optional)

## Step 3: Initialize and Apply the Configuration

Initialize Terraform:

```bash
terraform init
```

Apply the configuration:

```bash
terraform apply
```

Review the changes and type `yes` to create the VM.

## Step 4: Manage Your VM

### Querying VM Information

You can use the Firecracker data source to query information about existing VMs:

```hcl
data "firecracker_vm" "example_info" {
  vm_id = firecracker_vm.example.id
}

output "vm_details" {
  value = data.firecracker_vm.example_info
}
```

Apply the configuration to see the output:

```bash
terraform apply
```

### Managing Multiple VMs

You can create multiple VMs in the same configuration:

```hcl
resource "firecracker_vm" "vm1" {
  kernel_image_path = "/path/to/vmlinux"
  boot_args         = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"

  drives {
    drive_id       = "rootfs"
    path_on_host   = "/path/to/rootfs.ext4"
    is_root_device = true
    is_read_only   = false
  }

  machine_config {
    vcpu_count   = 1
    mem_size_mib = 512
  }
}

resource "firecracker_vm" "vm2" {
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
  }
}
```

### VM Lifecycle Management

Terraform manages the entire lifecycle of your VMs:

- **Create**: `terraform apply` creates new VMs
- **Update**: `terraform apply` updates existing VMs (note that most changes require recreation)
- **Destroy**: `terraform destroy` removes VMs

## Step 5: Verify the VM

You can verify that the VM was created by using the Firecracker API:

```bash
# Check machine configuration
curl http://localhost:8080/machine-config

# Check boot source
curl http://localhost:8080/boot-source

# Check drives
curl http://localhost:8080/drives
```

## Step 6: Clean Up

When you're done, you can destroy the VM:

```bash
terraform destroy
```

And tear down the Firecracker environment:

```bash
make teardown
```

## Advanced Usage

### Adding Multiple Drives

```hcl
resource "firecracker_vm" "multi_drive" {
  kernel_image_path = "/path/to/vmlinux"
  boot_args         = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"

  # Root filesystem
  drives {
    drive_id       = "rootfs"
    path_on_host   = "/path/to/rootfs.ext4"
    is_root_device = true
    is_read_only   = false
  }

  # Data volume
  drives {
    drive_id       = "data"
    path_on_host   = "/path/to/data.ext4"
    is_root_device = false
    is_read_only   = false
  }

  machine_config {
    vcpu_count   = 2
    mem_size_mib = 1024
  }
}
```

### Configuring Multiple Network Interfaces

```hcl
resource "firecracker_vm" "multi_network" {
  kernel_image_path = "/path/to/vmlinux"
  boot_args         = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw ip=dhcp"

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

  # Primary network
  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"
    guest_mac     = "AA:BB:CC:DD:EE:01"
  }

  # Secondary network
  network_interfaces {
    iface_id      = "eth1"
    host_dev_name = "tap1"
    guest_mac     = "AA:BB:CC:DD:EE:02"
  }
}
```

## Next Steps

- Learn more about the [firecracker_vm resource](../resources/vm.md)
- Explore the [firecracker_vm data source](../data-sources/vm.md)
- Check out the [example configurations](../../examples/)
- Set up [networking for your VMs](https://github.com/firecracker-microvm/firecracker/blob/main/docs/network-setup.md)
- Read the [Troubleshooting Guide](troubleshooting.md) if you encounter issues
