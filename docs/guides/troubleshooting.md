# Troubleshooting the Firecracker Terraform Provider

This guide provides solutions to common issues you might encounter when using the Firecracker Terraform Provider.

## Quick Diagnostics

The provider includes a built-in diagnostic tool to check the environment:

```bash
# Check the status of all components
make status

# Restart the environment
make teardown
make setup
```

## Common Issues

### Connection Refused

**Symptom:** You see an error like `connection refused` when Terraform tries to connect to the Firecracker API.

**Possible Causes:**
1. Firecracker is not running
2. The socat forwarding is not set up correctly
3. The `base_url` in your provider configuration is incorrect
4. A firewall is blocking the connection

**Solutions:**
1. Check the environment status:
   ```bash
   make status
   ```

2. Ensure Firecracker is running:
   ```bash
   ps aux | grep firecracker
   ```
   If not running, start it:
   ```bash
   make start-firecracker
   ```

3. Check if socat is running and forwarding correctly:
   ```bash
   ps aux | grep socat
   ```
   If not running, start it:
   ```bash
   make start-socat
   ```

4. Verify you can connect to the API manually:
   ```bash
   curl http://localhost:8080
   ```

5. Check for firewall issues:
   ```bash
   sudo iptables -L | grep 8080
   ```

### Permission Denied

**Symptom:** You see errors related to permissions when Firecracker tries to access kernel or rootfs images or /dev/kvm.

**Solutions:**

1. Ensure the user running Firecracker has read access to the kernel and rootfs images:
   ```bash
   chmod 644 /path/to/vmlinux
   chmod 644 /path/to/rootfs.ext4
   ```

2. Check KVM permissions:
   ```bash
   ls -la /dev/kvm
   ```
   
   If you don't have access:
   ```bash
   sudo chmod 666 /dev/kvm
   # Or add your user to the kvm group
   sudo usermod -aG kvm $(whoami)
   ```

3. Check socket permissions:
   ```bash
   ls -la /tmp/firecracker.sock
   ```
   
   If permissions are too restrictive:
   ```bash
   sudo chmod 666 /tmp/firecracker.sock
   ```

### Invalid Kernel or Rootfs Path

**Symptom:** Firecracker fails to start the VM with errors about invalid kernel or rootfs paths.

**Solutions:**
1. Verify the paths are correct and the files exist:
   ```bash
   ls -la /path/to/vmlinux
   ls -la /path/to/rootfs.ext4
   ```

2. Ensure you're using absolute paths in your Terraform configuration.

3. Check that the paths are accessible from the Firecracker process:
   ```bash
   sudo -u $(ps -o user= -p $(pgrep firecracker)) ls -la /path/to/vmlinux
   ```

4. If using the test configuration, update the paths in `test/main.tf` to match your environment.

### Network Interface Issues

**Symptom:** The VM starts but has no network connectivity.

**Solutions:**
1. Ensure the TAP device exists on the host:
   ```bash
   ip link show tap0
   ```

2. If it doesn't exist, create it:
   ```bash
   make setup-network
   ```
   
   Or manually:
   ```bash
   sudo ip tuntap add dev tap0 mode tap
   sudo ip link set tap0 up
   sudo ip addr add 172.16.0.1/24 dev tap0
   ```

3. Check IP forwarding is enabled:
   ```bash
   cat /proc/sys/net/ipv4/ip_forward
   ```
   
   If it's not enabled (output is 0):
   ```bash
   sudo sysctl -w net.ipv4.ip_forward=1
   ```

4. Verify NAT is set up correctly:
   ```bash
   sudo iptables -t nat -L POSTROUTING
   sudo iptables -L FORWARD
   ```

5. Test connectivity from inside the VM (requires console access):
   ```bash
   ping 172.16.0.1  # Should reach the host
   ping 8.8.8.8     # Should reach the internet if NAT is working
   ```

### VM Not Found During Import

**Symptom:** When trying to import a VM, you get an error that the VM was not found.

**Solutions:**
1. Verify the VM ID is correct and the VM exists:
   ```bash
   curl http://localhost:8080/machine-config
   ```

2. Check if Firecracker is running and the API is accessible:
   ```bash
   make status
   ```

3. Try restarting the environment:
   ```bash
   make teardown
   make setup
   ```

### Resource Already Exists

**Symptom:** When applying a configuration, you get an error that the resource already exists.

**Solution:**
1. Remove the resource from Terraform state:
   ```bash
   terraform state rm firecracker_vm.example
   ```

2. Stop and remove the existing VM:
   ```bash
   make teardown
   ```

3. Apply the configuration again:
   ```bash
   terraform apply
   ```

## Debugging Techniques

### Enable Terraform Logs

Set the `TF_LOG` environment variable to enable detailed logging:

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
terraform apply
```

### Check Firecracker Logs

If you started Firecracker with the `--log-path` option, check those logs:

```bash
cat /path/to/firecracker.log
```

Or start Firecracker with logging enabled:

```bash
firecracker --api-sock /tmp/firecracker.sock --log-path /tmp/firecracker.log --level Debug &
```

### API Request Debugging

You can use tools like `curl` to manually test the Firecracker API:

```bash
# Get machine configuration
curl http://localhost:8080/machine-config

# Get boot source configuration
curl http://localhost:8080/boot-source

# Create a VM (simplified example)
curl -X PUT -H "Content-Type: application/json" -d '{
  "kernel_image_path": "/path/to/vmlinux",
  "boot_args": "console=ttyS0"
}' http://localhost:8080/boot-source

# Configure machine
curl -X PUT -H "Content-Type: application/json" -d '{
  "vcpu_count": 2,
  "mem_size_mib": 1024
}' http://localhost:8080/machine-config

# Add a drive
curl -X PUT -H "Content-Type: application/json" -d '{
  "drive_id": "rootfs",
  "path_on_host": "/path/to/rootfs.ext4",
  "is_root_device": true,
  "is_read_only": false
}' http://localhost:8080/drives/rootfs

# Start the VM
curl -X PUT -H "Content-Type: application/json" -d '{
  "action_type": "InstanceStart"
}' http://localhost:8080/actions
```

### Inspecting the Provider

To debug the provider itself:

1. Build with debug symbols:
   ```bash
   go build -gcflags="all=-N -l" -o terraform-provider-firecracker
   ```

2. Use Delve for debugging:
   ```bash
   dlv exec ./terraform-provider-firecracker -- -debug
   ```

## Getting Help

If you continue to experience issues:

1. Run the diagnostic tool and share the output:
   ```bash
   make status > status.log
   ```

2. Check the [Firecracker documentation](https://github.com/firecracker-microvm/firecracker/tree/main/docs)

3. Open an issue on the [GitHub repository](https://github.com/terraform-provider-firecracker/terraform-provider-firecracker/issues) with:
   - A clear description of the issue
   - Steps to reproduce
   - Terraform and provider versions
   - Logs from Terraform and Firecracker
   - Output from `make status`

4. Join the [Firecracker community](https://github.com/firecracker-microvm/firecracker#community)
