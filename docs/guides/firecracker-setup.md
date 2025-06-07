# Setting Up Firecracker for Use with Terraform

This guide explains how to set up Firecracker to work with the Terraform provider.

## Prerequisites

- A Linux system (Firecracker only runs on Linux)
- Root or sudo access
- KVM virtualization support

## Step 1: Check KVM Support

Verify that your system supports KVM virtualization:

```bash
lscpu | grep Virtualization
```

You should see output indicating that virtualization is supported.

Also, check if the KVM module is loaded:

```bash
lsmod | grep kvm
```

If it's not loaded, load it:

```bash
sudo modprobe kvm
sudo modprobe kvm_intel  # For Intel processors
# OR
sudo modprobe kvm_amd    # For AMD processors
```

## Step 2: Install Firecracker

1. Download the latest Firecracker binary from the [releases page](https://github.com/firecracker-microvm/firecracker/releases):

   ```bash
   wget https://github.com/firecracker-microvm/firecracker/releases/download/v1.4.0/firecracker-v1.4.0-x86_64.tgz
   ```

   Replace the version number with the latest available version.

2. Extract the archive:

   ```bash
   tar -xvf firecracker-v1.4.0-x86_64.tgz
   ```

3. Move the binary to a location in your PATH:

   ```bash
   sudo mv release-v1.4.0-x86_64/firecracker-v1.4.0-x86_64 /usr/local/bin/firecracker
   sudo chmod +x /usr/local/bin/firecracker
   ```

4. Verify the installation:

   ```bash
   firecracker --version
   ```

## Step 3: Install socat

The Terraform provider communicates with Firecracker over HTTP, but Firecracker uses a Unix socket. We'll use `socat` to bridge this gap:

```bash
sudo apt-get update
sudo apt-get install -y socat  # For Debian/Ubuntu
# OR
sudo yum install -y socat      # For CentOS/RHEL
```

## Step 4: Prepare Kernel and Root Filesystem

You need a Linux kernel and a root filesystem to create Firecracker VMs:

### Option 1: Use Pre-built Images

1. Download a pre-built kernel and rootfs:

   ```bash
   wget https://s3.amazonaws.com/spec.ccfc.min/img/hello/kernel/hello-vmlinux.bin
   wget https://s3.amazonaws.com/spec.ccfc.min/img/hello/fsfiles/hello-rootfs.ext4
   ```

### Option 2: Build Your Own

1. For the kernel, follow the [Firecracker kernel setup guide](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md).

2. For the root filesystem, you can create a custom one or use a minimal distribution like Alpine Linux.

## Step 5: Set Up Network (Optional)

If you want your VMs to have network connectivity:

1. Create a TAP device:

   ```bash
   sudo ip tuntap add dev tap0 mode tap
   sudo ip addr add 172.16.0.1/24 dev tap0
   sudo ip link set tap0 up
   ```

2. Set up IP forwarding and NAT:

   ```bash
   sudo sh -c "echo 1 > /proc/sys/net/ipv4/ip_forward"
   sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
   sudo iptables -A FORWARD -i eth0 -o tap0 -j ACCEPT
   sudo iptables -A FORWARD -i tap0 -o eth0 -j ACCEPT
   ```

   Replace `eth0` with your actual internet-facing interface.

## Step 6: Start Firecracker with API Access

1. Create a socket for Firecracker:

   ```bash
   rm -f /tmp/firecracker.sock
   ```

2. Start Firecracker:

   ```bash
   firecracker --api-sock /tmp/firecracker.sock &
   ```

3. Use socat to forward TCP traffic to the Unix socket:

   ```bash
   socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock &
   ```

   This makes the Firecracker API available at `http://localhost:8080`.

## Step 7: Verify the Setup

Test that you can communicate with the Firecracker API:

```bash
curl --unix-socket /tmp/firecracker.sock http://localhost/
# OR
curl http://localhost:8080/
```

You should receive a response from the Firecracker API.

## Next Steps

Now that you have Firecracker set up, you can:

1. Configure the [Terraform provider](../index.md)
2. Create your first [Firecracker VM](../resources/vm.md)
3. Follow the [Getting Started guide](getting-started.md)

## Troubleshooting

### Permission Denied

If you see permission errors when accessing the socket:

```bash
sudo chmod 777 /tmp/firecracker.sock
```

### Connection Refused

If you can't connect to the API:

1. Check if Firecracker is running:
   ```bash
   ps aux | grep firecracker
   ```

2. Check if socat is running:
   ```bash
   ps aux | grep socat
   ```

3. Restart both if needed.

### Missing KVM Support

If Firecracker fails with KVM errors, ensure:

1. Your system supports virtualization
2. KVM modules are loaded
3. You have permission to access `/dev/kvm`:
   ```bash
   sudo chmod 666 /dev/kvm
   ```
