#!/bin/bash

# This script helps debug Firecracker VM boot issues

echo "Firecracker VM Debugging Tool"
echo "============================"

# Get paths from Terraform configuration
KERNEL_PATH=$(grep "kernel_image_path" test/main.tf | sed 's/.*kernel_image_path[^"]*"\([^"]*\)".*/\1/')
ROOTFS_PATH=$(grep "path_on_host" test/main.tf | head -1 | sed 's/.*path_on_host[^"]*"\([^"]*\)".*/\1/')

echo "Detected paths from Terraform config:"
echo "Kernel: $KERNEL_PATH"
echo "Rootfs: $ROOTFS_PATH"
echo ""

# Check if kernel image exists
if [ -f "$KERNEL_PATH" ]; then
  echo "[✓] Kernel image exists: $KERNEL_PATH"
  ls -la "$KERNEL_PATH"
else
  echo "[✗] Kernel image not found: $KERNEL_PATH"
fi

# Check if rootfs exists
if [ -f "$ROOTFS_PATH" ]; then
  echo "[✓] Root filesystem exists: $ROOTFS_PATH"
  ls -la "$ROOTFS_PATH"
  
  # Check filesystem type
  FSTYPE=$(file -b "$ROOTFS_PATH")
  echo "    Filesystem type: $FSTYPE"
  
  # Try to mount the filesystem to verify it's valid
  echo "    Attempting to mount filesystem to verify integrity..."
  MOUNT_DIR=$(mktemp -d)
  if sudo mount -o loop "$ROOTFS_PATH" "$MOUNT_DIR" 2>/dev/null; then
    echo "    [✓] Filesystem mounted successfully"
    ls -la "$MOUNT_DIR"
    sudo umount "$MOUNT_DIR"
  else
    echo "    [✗] Failed to mount filesystem - it may be corrupted"
  fi
  rmdir "$MOUNT_DIR"
else
  echo "[✗] Root filesystem not found: $ROOTFS_PATH"
fi

# Check if tap devices exist
TAP_DEVICE="tap0"
if ip link show "$TAP_DEVICE" >/dev/null 2>&1; then
  echo "[✓] TAP device exists: $TAP_DEVICE"
  ip link show "$TAP_DEVICE"
else
  echo "[✗] TAP device not found: $TAP_DEVICE"
  echo "    You can create it with: sudo ip tuntap add dev $TAP_DEVICE mode tap"
fi

echo ""
echo "Firecracker API Check"
echo "===================="
# Check if Firecracker API is running
if curl -s http://localhost:8080/machine-config >/dev/null 2>&1; then
  echo "[✓] Firecracker API is running"
else
  echo "[✗] Firecracker API is not running"
  echo "    Make sure the Firecracker process is started with API socket"
fi

echo ""
echo "Suggested Boot Arguments"
echo "======================"
echo "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda1 rootfstype=ext4 rw"
echo ""
echo "Debug complete!"
