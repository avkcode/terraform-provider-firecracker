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

resource "firecracker_vm" "multi_drive_example" {
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

  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"
  }
}
