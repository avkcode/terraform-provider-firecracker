# Multiple Drives Firecracker VM Example

This example demonstrates how to create a Firecracker microVM with multiple drives using the Terraform provider.

## Prerequisites

- Firecracker running with API access
- A Linux kernel binary (vmlinux format)
- Multiple filesystem images (ext4 format)

## Usage

1. Update the paths in `main.tf` to point to your actual kernel and filesystem files.

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

This example creates a Firecracker microVM with:

- 2 vCPUs
- 2048 MiB of memory
- Three drives:
  - A root filesystem (read-write)
  - A data volume (read-write)
  - A content volume (read-only)
- A network interface

## Notes

- Make sure Firecracker is running and accessible at http://localhost:8080
- All filesystem paths must be accessible by the Firecracker process
- Inside the VM, the drives will be available as `/dev/vda` (root), `/dev/vdb` (data), and `/dev/vdc` (content)
