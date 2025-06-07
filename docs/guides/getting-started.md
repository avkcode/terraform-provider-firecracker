# Getting Started with the Firecracker Terraform Provider

This guide will help you get started with the Firecracker Terraform Provider.

## Prerequisites

Before you begin, ensure you have the following:

1. [Terraform](https://www.terraform.io/downloads.html) >= 1.0
2. [Firecracker](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md) installed
3. A Linux kernel binary (vmlinux format)
4. A root filesystem image (ext4 format)

## Step 1: Set Up Firecracker

First, you need to set up Firecracker and expose its API:

```bash
# Create a socket for Firecracker
rm -f /tmp/firecracker.sock

# Start Firecracker
firecracker --api-sock /tmp/firecracker.sock &

# Use socat to forward TCP traffic to the Unix socket
socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock &
```

## Step 2: Install the Terraform Provider

Build and install the provider:

```bash
git clone https://github.com/avkcode/terraform-provider-firecracker.git
cd terraform-provider-firecracker
make build
```

## Step 3: Create Your Terraform Configuration

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
      source = "avkcode/firecracker"
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

## Step 4: Initialize and Apply the Configuration

Initialize Terraform:

```bash
terraform init
```

Apply the configuration:

```bash
terraform apply
```

Review the changes and type `yes` to create the VM.

## Step 5: Verify the VM

You can verify that the VM was created by using the Firecracker API:

```bash
curl http://localhost:8080/vm
```

## Step 6: Clean Up

When you're done, you can destroy the VM:

```bash
terraform destroy
```

## Next Steps

- Learn more about the [firecracker_vm resource](../resources/vm.md)
- Explore the [Firecracker API documentation](https://github.com/firecracker-microvm/firecracker/blob/main/docs/api_requests)
- Set up [networking for your VMs](https://github.com/firecracker-microvm/firecracker/blob/main/docs/network-setup.md)
