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

resource "firecracker_vm" "networking_example" {
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
