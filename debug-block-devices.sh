#!/bin/bash

# This script helps debug block device issues in Firecracker VMs

echo "Firecracker Block Device Debugging Tool"
echo "======================================"

# Check if we're running inside a Firecracker VM
if [ -f /sys/class/dmi/id/product_name ]; then
  PRODUCT_NAME=$(cat /sys/class/dmi/id/product_name)
  if [[ "$PRODUCT_NAME" == *"Firecracker"* ]]; then
    echo "[✓] Running inside a Firecracker VM"
  else
    echo "[✗] Not running inside a Firecracker VM"
    echo "    This script is meant to be run inside the VM to debug block devices"
    exit 1
  fi
else
  echo "[✗] Cannot determine if running inside a Firecracker VM"
  echo "    This script is meant to be run inside the VM to debug block devices"
  exit 1
fi

# List all block devices
echo ""
echo "Block Devices"
echo "============"
lsblk -a
echo ""

# Check for vda and partitions
if [ -b /dev/vda ]; then
  echo "[✓] /dev/vda exists"
  ls -la /dev/vda*
  
  # Check if vda has partitions
  if [ -b /dev/vda1 ]; then
    echo "[✓] /dev/vda1 exists"
    
    # Try to mount vda1
    echo "Attempting to mount /dev/vda1..."
    MOUNT_DIR=$(mktemp -d)
    if mount /dev/vda1 $MOUNT_DIR 2>/dev/null; then
      echo "[✓] Successfully mounted /dev/vda1"
      ls -la $MOUNT_DIR
      umount $MOUNT_DIR
    else
      echo "[✗] Failed to mount /dev/vda1"
    fi
    rmdir $MOUNT_DIR
  else
    echo "[✗] /dev/vda1 does not exist"
    
    # Check if vda is partitioned
    echo "Checking partition table on /dev/vda..."
    fdisk -l /dev/vda
    
    # Try to mount vda directly
    echo "Attempting to mount /dev/vda directly..."
    MOUNT_DIR=$(mktemp -d)
    if mount /dev/vda $MOUNT_DIR 2>/dev/null; then
      echo "[✓] Successfully mounted /dev/vda"
      ls -la $MOUNT_DIR
      umount $MOUNT_DIR
    else
      echo "[✗] Failed to mount /dev/vda"
    fi
    rmdir $MOUNT_DIR
  fi
else
  echo "[✗] /dev/vda does not exist"
  echo "Available block devices:"
  find /dev -type b | sort
fi

echo ""
echo "Kernel Command Line"
echo "=================="
cat /proc/cmdline

echo ""
echo "Debug complete!"
