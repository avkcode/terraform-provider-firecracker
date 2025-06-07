# Configure the Firecracker provider
provider "firecracker" {
  base_url = "http://localhost:8080" # Replace with the actual base URL of your Firecracker API
}

# Define a Firecracker VM resource
resource "firecracker_vm" "example_vm" {
  kernel_image_path = "/srv/terraform-provider-firecracker/test/vmlinux" # Path to the kernel image
  boot_args         = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"

  drives {
    drive_id       = "rootfs"
    path_on_host   = "/srv/terraform-provider-firecracker/test/firecracker-rootfs.ext4" # Path to the root filesystem
    is_root_device = true
    is_read_only   = false
  }

  machine_config {
    vcpu_count   = 2          # Number of vCPUs
    mem_size_mib = 9096       # Memory size in MiB
  }

  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"    # Host device name for the interface
    guest_mac     = "06:00:00:00:00:01"
  }
}
