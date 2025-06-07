# Installing the Firecracker Terraform Provider

This guide explains how to install and configure the Firecracker Terraform Provider.

## Prerequisites

Before installing the provider, ensure you have:

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (only needed if building from source)
- [Firecracker](https://github.com/firecracker-microvm/firecracker/releases) binary installed and accessible in your PATH
- [socat](http://www.dest-unreach.org/socat/) for forwarding HTTP traffic to the Firecracker Unix socket
- Linux with KVM support (Firecracker only runs on Linux)
- Appropriate permissions to access `/dev/kvm` (typically requires membership in the `kvm` group)

You can verify your environment with:

```bash
# Check Terraform installation
terraform version

# Check Go installation (if building from source)
go version

# Check Firecracker installation
firecracker --version

# Check socat installation
socat -V

# Check KVM access
ls -la /dev/kvm
```

## Installation Methods

### Method 1: Build from Source (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/terraform-provider-firecracker/terraform-provider-firecracker.git
   cd terraform-provider-firecracker
   ```

2. Verify your environment meets all requirements:
   ```bash
   make verify
   ```
   
   This command checks for all required dependencies and tools.

3. Build and install the provider:
   ```bash
   make build
   ```

   This will build the provider and install it to your local Terraform plugin directory.

4. Set up the Firecracker environment:
   ```bash
   make setup
   ```
   
   This starts Firecracker and configures socat to forward traffic.

### Method 2: Manual Installation

1. Download the latest release from the [GitHub releases page](https://github.com/terraform-provider-firecracker/terraform-provider-firecracker/releases).

2. Extract the archive:
   ```bash
   unzip terraform-provider-firecracker_*.zip
   ```

3. Create the plugin directory (if it doesn't exist):
   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/terraform-provider-firecracker/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)
   ```

4. Move the provider binary to the plugin directory:
   ```bash
   mv terraform-provider-firecracker ~/.terraform.d/plugins/registry.terraform.io/terraform-provider-firecracker/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)/
   ```

5. Start Firecracker and socat manually:
   ```bash
   # Remove any existing socket
   rm -f /tmp/firecracker.sock
   
   # Start Firecracker
   firecracker --api-sock /tmp/firecracker.sock &
   
   # Start socat to forward traffic
   socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock &
   ```

## Makefile Utilities

The provider includes a comprehensive Makefile with useful commands:

```bash
# Show all available commands
make help

# Check all dependencies
make verify

# Build the provider
make build

# Set up the Firecracker environment
make setup

# Check the status of all components
make status

# Tear down the environment
make teardown

# Clean up temporary files
make clean

# Run a test configuration
make run
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
         source = "terraform-provider-firecracker/firecracker"
         version = "0.1.0"
       }
     }
   }

   provider "firecracker" {
     base_url = "http://localhost:8080"
   }
   
   # Optional: Create a simple VM to test the provider
   resource "firecracker_vm" "test" {
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
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

4. Verify the provider is properly installed:
   ```bash
   terraform providers
   ```

   You should see the Firecracker provider listed.

5. If you included the VM resource, you can apply the configuration:
   ```bash
   terraform apply
   ```

## Troubleshooting

### Provider Not Found

If Terraform cannot find the provider, ensure:

1. The provider binary is in the correct location:
   ```bash
   ls -la ~/.terraform.d/plugins/registry.terraform.io/terraform-provider-firecracker/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)/
   ```

2. The provider source in your Terraform configuration matches the directory structure:
   ```hcl
   terraform {
     required_providers {
       firecracker = {
         source = "terraform-provider-firecracker/firecracker"
         version = "0.1.0"
       }
     }
   }
   ```

3. You're using the correct version number.

4. Check the Terraform debug logs:
   ```bash
   TF_LOG=DEBUG terraform init
   ```

### Permission Issues

If you encounter permission issues:

```bash
# Make the provider executable
chmod +x ~/.terraform.d/plugins/registry.terraform.io/terraform-provider-firecracker/firecracker/0.1.0/$(go env GOOS)_$(go env GOARCH)/terraform-provider-firecracker

# Ensure you have access to /dev/kvm
sudo chmod 666 /dev/kvm
# Or add your user to the kvm group
sudo usermod -aG kvm $(whoami)
```

### Firecracker or socat Not Running

Check the status of the environment:

```bash
make status
```

If any component is not running, restart the environment:

```bash
make teardown
make setup
```

## Next Steps

Now that you have installed the Firecracker Terraform Provider, you can:

- Follow the [Getting Started Guide](getting-started.md)
- Learn about the [firecracker_vm resource](../resources/vm.md)
- Explore the [firecracker_vm data source](../data-sources/vm.md)
- Set up [Firecracker](firecracker-setup.md) if you haven't already
- Check the [Troubleshooting Guide](troubleshooting.md) if you encounter issues
