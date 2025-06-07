# firecracker_vm Resource

Manages a Firecracker microVM.

## Example Usage

### Basic VM

```hcl
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
    guest_mac     = "AA:BB:CC:DD:EE:FF"
  }
}
```

### VM with Multiple Drives

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

  # Read-only content
  drives {
    drive_id       = "content"
    path_on_host   = "/path/to/content.ext4"
    is_root_device = false
    is_read_only   = true
  }

  machine_config {
    vcpu_count   = 2
    mem_size_mib = 2048
  }
}
```

### VM with Multiple Network Interfaces

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

  # Primary network interface
  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"
    guest_mac     = "AA:BB:CC:DD:EE:01"
  }

  # Secondary network interface
  network_interfaces {
    iface_id      = "eth1"
    host_dev_name = "tap1"
    guest_mac     = "AA:BB:CC:DD:EE:02"
  }
}
```

## Argument Reference

### Required Arguments

* `kernel_image_path` - (Required) Path to the kernel image. Must be accessible by the Firecracker process. This should be an uncompressed Linux kernel binary (vmlinux format).
* `drives` - (Required) List of drives attached to the VM. At least one drive must be specified, typically containing the root filesystem.
* `machine_config` - (Required) Machine configuration for the VM. This defines the virtual hardware resources allocated to the VM.

### Optional Arguments

* `boot_args` - (Optional) Boot arguments for the kernel. Default is `console=ttyS0 noapic reboot=k panic=1 pci=off root=/dev/vda rootfstype=ext4 rw init=/sbin/init`.
* `network_interfaces` - (Optional) List of network interfaces attached to the VM. Each interface connects to a TAP device on the host.

### `drives` Block Arguments

* `drive_id` - (Required) ID of the drive. This is used to identify the drive within Firecracker and must be unique within the VM.
* `path_on_host` - (Required) Path to the drive on the host. This must be accessible by the Firecracker process and should be a valid disk image (e.g., ext4 filesystem).
* `is_root_device` - (Required) Whether this drive is the root device. Only one drive can be marked as the root device.
* `is_read_only` - (Optional) Whether the drive is read-only. Default is `false`.

### `machine_config` Block Arguments

* `vcpu_count` - (Required) Number of vCPUs. Must be between 1 and 32.
* `mem_size_mib` - (Required) Memory size in MiB. Must be between 128 and 32768.

### `network_interfaces` Block Arguments

* `iface_id` - (Required) ID of the network interface. This is used to identify the interface within Firecracker and must be unique within the VM.
* `host_dev_name` - (Required) Host device name for the interface. This should be a TAP device that exists on the host (e.g., 'tap0').
* `guest_mac` - (Optional) MAC address for the guest network interface. If not specified, Firecracker will generate one. Format: 'XX:XX:XX:XX:XX:XX'.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The ID of the VM.

## Timeouts

The `timeouts` block allows you to specify how long certain operations are allowed to take before being considered an error:

```hcl
resource "firecracker_vm" "example" {
  # ... other configuration ...

  timeouts {
    create = "10m"
    update = "5m"
    delete = "5m"
  }
}
```

* `create` - (Default `10m`) How long to wait for the VM to be created.
* `update` - (Default `5m`) How long to wait for the VM to be updated.
* `delete` - (Default `5m`) How long to wait for the VM to be deleted.

## Update Behavior

Most changes to a Firecracker VM require recreation of the VM. This is because Firecracker does not support modifying most VM properties after creation. The following changes will trigger recreation:

* Changes to `kernel_image_path`
* Changes to `boot_args`
* Changes to `drives` configuration
* Changes to `machine_config`
* Changes to `network_interfaces`

## Using with Provisioners

You can use Terraform provisioners with Firecracker VMs if your VM has network connectivity and SSH access:

```hcl
resource "firecracker_vm" "example" {
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

  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"
  }

  # SSH connection for provisioning
  connection {
    type        = "ssh"
    user        = "root"
    private_key = file("~/.ssh/id_rsa")
    host        = "172.16.0.2"  # IP address of the VM
  }

  # File provisioner example
  provisioner "file" {
    source      = "local/path/to/file"
    destination = "/remote/path/on/vm"
  }

  # Remote-exec provisioner example
  provisioner "remote-exec" {
    inline = [
      "apt-get update",
      "apt-get install -y nginx",
      "systemctl start nginx"
    ]
  }
}
```

> **Note:** For provisioners to work, your VM must have:
> 1. Network connectivity (properly configured TAP device)
> 2. SSH server installed and running
> 3. Proper firewall rules to allow SSH connections

## Resource Dependencies

You can create dependencies between Firecracker VMs and other resources:

```hcl
# Create a TAP device first
resource "null_resource" "tap_device" {
  provisioner "local-exec" {
    command = "sudo ip tuntap add dev tap0 mode tap && sudo ip link set tap0 up && sudo ip addr add 172.16.0.1/24 dev tap0"
  }
}

# Then create a VM that depends on the TAP device
resource "firecracker_vm" "example" {
  depends_on = [null_resource.tap_device]

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

## Import

Firecracker VMs can be imported using the VM ID:

```bash
terraform import firecracker_vm.example <vm-id>
```

This allows you to bring existing Firecracker VMs under Terraform management.
