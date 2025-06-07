# Basic Firecracker VM Example

This example demonstrates how to create a basic Firecracker microVM using the Terraform provider.

## Prerequisites

- Firecracker running with API access
- A Linux kernel binary (vmlinux format)
- A root filesystem image (ext4 format)

## Usage

1. Update the paths in `main.tf` to point to your actual kernel and rootfs files.

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

4. When you're done, destroy the resources:
   ```bash
   terraform destroy
   ```

## What This Example Creates

This example creates a simple Firecracker microVM with:

- 2 vCPUs
- 1024 MiB of memory
- A single root drive
- No network interfaces

## Notes

- Make sure Firecracker is running and accessible at http://localhost:8080
- The kernel and rootfs paths must be accessible by the Firecracker process
