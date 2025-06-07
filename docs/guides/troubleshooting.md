# Troubleshooting the Firecracker Terraform Provider

This guide provides solutions to common issues you might encounter when using the Firecracker Terraform Provider.

## Common Issues

### Connection Refused

**Symptom:** You see an error like `connection refused` when Terraform tries to connect to the Firecracker API.

**Possible Causes:**
1. Firecracker is not running
2. The socat forwarding is not set up correctly
3. The `base_url` in your provider configuration is incorrect

**Solutions:**
1. Ensure Firecracker is running:
   ```bash
   ps aux | grep firecracker
   ```
2. Check if socat is running and forwarding correctly:
   ```bash
   ps aux | grep socat
   ```
3. Verify you can connect to the API manually:
   ```bash
   curl http://localhost:8080
   ```

### Permission Denied

**Symptom:** You see errors related to permissions when Firecracker tries to access kernel or rootfs images.

**Solution:**
Ensure the user running Firecracker has read access to the kernel and rootfs images:
```bash
chmod 644 /path/to/vmlinux
chmod 644 /path/to/rootfs.ext4
```

### Invalid Kernel or Rootfs Path

**Symptom:** Firecracker fails to start the VM with errors about invalid kernel or rootfs paths.

**Solution:**
1. Verify the paths are correct and the files exist
2. Ensure you're using absolute paths
3. Check that the paths are accessible from the Firecracker process

### Network Interface Issues

**Symptom:** The VM starts but has no network connectivity.

**Solutions:**
1. Ensure the TAP device exists on the host:
   ```bash
   ip link show tap0
   ```
2. If it doesn't exist, create it:
   ```bash
   sudo ip tuntap add dev tap0 mode tap
   sudo ip link set tap0 up
   sudo ip addr add 172.16.0.1/24 dev tap0
   ```

### VM Not Found During Import

**Symptom:** When trying to import a VM, you get an error that the VM was not found.

**Solution:**
Verify the VM ID is correct and the VM exists:
```bash
curl http://localhost:8080/vm/your-vm-id
```

## Debugging

### Enable Terraform Logs

Set the `TF_LOG` environment variable to enable detailed logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

### Check Firecracker Logs

If you started Firecracker with the `--log-path` option, check those logs:

```bash
cat /path/to/firecracker.log
```

### API Request Debugging

You can use tools like `curl` to manually test the Firecracker API:

```bash
# Get VM info
curl http://localhost:8080/vm/your-vm-id

# Create a VM (simplified example)
curl -X PUT -H "Content-Type: application/json" -d '{
  "boot-source": {
    "kernel_image_path": "/path/to/vmlinux",
    "boot_args": "console=ttyS0"
  }
}' http://localhost:8080/boot-source
```

## Getting Help

If you continue to experience issues:

1. Check the [Firecracker documentation](https://github.com/firecracker-microvm/firecracker/tree/main/docs)
2. Open an issue on the [GitHub repository](https://github.com/avkcode/terraform-provider-firecracker/issues)
3. Join the [Firecracker community](https://github.com/firecracker-microvm/firecracker#community)
