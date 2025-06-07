terraform {
  required_providers {
    firecracker = {
      source = "hashicorp/firecracker"
      version = "0.1.0"
    }
  }
}

provider "firecracker" {
  base_url = "http://localhost:8080"
}

resource "firecracker_vm" "docker_vm" {
  kernel_image_path = "../vmlinux"
  boot_args         = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw ip=dhcp"

  drives {
    drive_id       = "rootfs"
    path_on_host   = "../firecracker-rootfs.ext4"
    is_root_device = true
    is_read_only   = false
  }

  machine_config {
    vcpu_count   = 2
    mem_size_mib = 2048
  }

  network_interfaces {
    iface_id      = "eth0"
    host_dev_name = "tap0"
    guest_mac     = "AA:BB:CC:DD:EE:01"
  }

  # SSH connection for remote-exec
  connection {
    type        = "ssh"
    user        = "root"
    private_key = file("../id_rsa")
    host        = "172.16.0.2"  # Assuming this is the IP of the VM
    timeout     = "2m"
  }

  # Install Docker using remote-exec
  provisioner "remote-exec" {
    inline = [
      "echo 'Installing Docker...'",
      "apt-get update",
      "apt-get install -y apt-transport-https ca-certificates curl software-properties-common",
      "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -",
      "add-apt-repository 'deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable'",
      "apt-get update",
      "apt-get install -y docker-ce",
      "systemctl enable docker",
      "systemctl start docker",
      "docker --version",
      "echo 'Docker installation complete!'"
    ]
  }
}
