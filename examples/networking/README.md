# Networking Firecracker VM Example

This example demonstrates how to create a Firecracker microVM with multiple network interfaces using the Terraform provider.

## Prerequisites

- Firecracker running with API access
- A Linux kernel binary (vmlinux format)
- A root filesystem image (ext4 format)
- TAP devices set up on the host

## Setting Up TAP Devices

Before running this example, set up the TAP devices on your host:

```bash
# Create the first TAP device
sudo ip tuntap add dev tap0 mode tap
sudo ip addr add 172.16.0.1/24 dev tap0
sudo ip link set tap0 up

# Create the second TAP device
sudo ip tuntap add dev tap1 mode tap
sudo ip addr add 172.17.0.1/24 dev tap1
sudo ip link set tap1 up

# Enable IP forwarding
sudo sh -c "echo 1 > /proc/sys/net/ipv4/ip_forward"

# Set up NAT for internet access
sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
sudo iptables -A FORWARD -i eth0 -o tap0 -j ACCEPT
sudo iptables -A FORWARD -i tap0 -o eth0 -j ACCEPT
sudo iptables -A FORWARD -i eth0 -o tap1 -j ACCEPT
sudo iptables -A FORWARD -i tap1 -o eth0 -j ACCEPT
```

Replace `eth0` with your actual internet-facing interface.

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

This example creates a Firecracker microVM with:

- 2 vCPUs
- 1024 MiB of memory
- A root filesystem
- Two network interfaces:
  - eth0 connected to tap0 (172.16.0.0/24 network)
  - eth1 connected to tap1 (172.17.0.0/24 network)

## Notes

- Make sure Firecracker is running and accessible at http://localhost:8080
- The kernel and rootfs paths must be accessible by the Firecracker process
- The boot arguments include `ip=dhcp` to enable DHCP for network configuration
- You may need to set up a DHCP server on your host to provide IP addresses to the VM
