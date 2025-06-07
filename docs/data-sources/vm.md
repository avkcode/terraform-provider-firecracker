# firecracker_vm Data Source

Use this data source to retrieve information about an existing Firecracker microVM.

## Example Usage

```hcl
data "firecracker_vm" "example" {
  vm_id = "my-vm-id"
}

output "kernel_path" {
  value = data.firecracker_vm.example.kernel_image_path
}

output "vcpu_count" {
  value = data.firecracker_vm.example.machine_config[0].vcpu_count
}
```

## Argument Reference

* `vm_id` - (Required) ID of the Firecracker VM to retrieve information about.

## Attributes Reference

In addition to the argument above, the following attributes are exported:

* `kernel_image_path` - Path to the kernel image.
* `boot_args` - Boot arguments for the kernel.
* `drives` - List of drives attached to the VM.
  * `drive_id` - ID of the drive.
  * `path_on_host` - Path to the drive on the host.
  * `is_root_device` - Whether this drive is the root device.
  * `is_read_only` - Whether the drive is read-only.
* `machine_config` - Machine configuration for the VM.
  * `vcpu_count` - Number of vCPUs.
  * `mem_size_mib` - Memory size in MiB.
* `network_interfaces` - List of network interfaces attached to the VM.
  * `iface_id` - ID of the network interface.
  * `host_dev_name` - Host device name for the interface.
  * `guest_mac` - MAC address for the guest network interface.
