# Installing the Firecracker Terraform Provider

This guide explains how to install the Firecracker Terraform Provider.

## Prerequisites

Before installing the provider, ensure you have:

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (only needed if building from source)

## Installation Methods

### Method 1: Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/avkcode/terraform-provider-firecracker.git
   cd terraform-provider-firecracker
   ```

2. Build and install the provider:
   ```bash
   make build
   ```

   This will build the provider and install it to your local Terraform plugin directory.

### Method 2: Manual Installation

1. Download the latest release from the [GitHub releases page](https://github.com/avkcode/terraform-provider-firecracker/releases).

2. Extract the archive:
   ```bash
   unzip terraform-provider-firecracker_*.zip
   ```

3. Create the plugin directory (if it doesn't exist):
   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/avkcode/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)
   ```

4. Move the provider binary to the plugin directory:
   ```bash
   mv terraform-provider-firecracker ~/.terraform.d/plugins/registry.terraform.io/avkcode/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)/
   ```

## Verifying the Installation

Create a simple Terraform configuration to verify the installation:

1. Create a new directory:
   ```bash
   mkdir terraform-firecracker-test
   cd terraform-firecracker-test
   ```

2. Create a file named `main.tf` with the following content:
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
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

   If the installation was successful, you should see a message indicating that the provider was successfully initialized.

## Troubleshooting

### Provider Not Found

If Terraform cannot find the provider, ensure:

1. The provider binary is in the correct location
2. The provider source in your Terraform configuration matches the directory structure
3. You're using the correct version number

### Permission Issues

If you encounter permission issues:

```bash
chmod +x ~/.terraform.d/plugins/registry.terraform.io/avkcode/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)/terraform-provider-firecracker
```

### Version Mismatch

If you see a version mismatch error, ensure the version in your Terraform configuration matches the installed version.

## Next Steps

Now that you have installed the Firecracker Terraform Provider, you can:

- Follow the [Getting Started Guide](getting-started.md)
- Learn about the [firecracker_vm resource](../resources/vm.md)
- Set up [Firecracker](firecracker-setup.md) if you haven't already
