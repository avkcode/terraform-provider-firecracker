package firecracker

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFirecrackerVM() *schema.Resource {
    return &schema.Resource{
        ReadContext: dataSourceFirecrackerVMRead,
        Schema: map[string]*schema.Schema{
            "vm_id": {
                Type:        schema.TypeString,
                Required:    true,
                Description: "ID of the Firecracker VM to retrieve information about.",
            },
            "kernel_image_path": {
                Type:        schema.TypeString,
                Computed:    true,
                Description: "Path to the kernel image used by the VM.",
            },
            "boot_args": {
                Type:        schema.TypeString,
                Computed:    true,
                Description: "Boot arguments for the kernel.",
            },
            "drives": {
                Type:        schema.TypeList,
                Computed:    true,
                Description: "List of drives attached to the VM.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "drive_id": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "ID of the drive.",
                        },
                        "path_on_host": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "Path to the drive on the host.",
                        },
                        "is_root_device": {
                            Type:        schema.TypeBool,
                            Computed:    true,
                            Description: "Whether this drive is the root device.",
                        },
                        "is_read_only": {
                            Type:        schema.TypeBool,
                            Computed:    true,
                            Description: "Whether the drive is read-only.",
                        },
                    },
                },
            },
            "machine_config": {
                Type:        schema.TypeList,
                Computed:    true,
                Description: "Machine configuration for the VM.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "vcpu_count": {
                            Type:        schema.TypeInt,
                            Computed:    true,
                            Description: "Number of vCPUs.",
                        },
                        "mem_size_mib": {
                            Type:        schema.TypeInt,
                            Computed:    true,
                            Description: "Memory size in MiB.",
                        },
                    },
                },
            },
            "network_interfaces": {
                Type:        schema.TypeList,
                Computed:    true,
                Description: "List of network interfaces attached to the VM.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "iface_id": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "ID of the network interface.",
                        },
                        "host_dev_name": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "Host device name for the interface.",
                        },
                        "guest_mac": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "MAC address for the guest.",
                        },
                    },
                },
            },
        },
        Timeouts: &schema.ResourceTimeout{
            Read: schema.DefaultTimeout(1 * time.Minute),
        },
    }
}

func dataSourceFirecrackerVMRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)
    var diags diag.Diagnostics

    vmID := d.Get("vm_id").(string)
    tflog.Debug(ctx, "Reading Firecracker VM for data source", map[string]interface{}{
        "id": vmID,
    })

    // Get VM details from the API
    vmInfo, err := client.GetVM(ctx, vmID)
    if err != nil {
        return diag.FromErr(fmt.Errorf("error reading VM for data source: %w", err))
    }

    // If VM not found, return error
    if vmInfo == nil {
        return diag.FromErr(fmt.Errorf("VM with ID %s not found", vmID))
    }

    // Set the ID
    d.SetId(vmID)

    // Update the resource data based on the VM info
    if bootSource, ok := vmInfo["boot-source"].(map[string]interface{}); ok {
        if kernelPath, ok := bootSource["kernel_image_path"].(string); ok {
            d.Set("kernel_image_path", kernelPath)
        }
        if bootArgs, ok := bootSource["boot_args"].(string); ok {
            d.Set("boot_args", bootArgs)
        }
    }

    // Handle machine config
    if machineConfig, ok := vmInfo["machine-config"].(map[string]interface{}); ok {
        newMachineConfig := []map[string]interface{}{
            {
                "vcpu_count":   machineConfig["vcpu_count"],
                "mem_size_mib": machineConfig["mem_size_mib"],
            },
        }
        d.Set("machine_config", newMachineConfig)
    }

    // Handle drives
    if drives, ok := vmInfo["drives"].([]interface{}); ok {
        newDrives := make([]map[string]interface{}, 0, len(drives))
        for _, driveRaw := range drives {
            if drive, ok := driveRaw.(map[string]interface{}); ok {
                newDrive := map[string]interface{}{
                    "drive_id":       drive["drive_id"],
                    "path_on_host":   drive["path_on_host"],
                    "is_root_device": drive["is_root_device"],
                    "is_read_only":   drive["is_read_only"],
                }
                newDrives = append(newDrives, newDrive)
            }
        }
        d.Set("drives", newDrives)
    }

    // Handle network interfaces
    if networkInterfaces, ok := vmInfo["network-interfaces"].([]interface{}); ok {
        newInterfaces := make([]map[string]interface{}, 0, len(networkInterfaces))
        for _, ifaceRaw := range networkInterfaces {
            if iface, ok := ifaceRaw.(map[string]interface{}); ok {
                newIface := map[string]interface{}{
                    "iface_id":      iface["iface_id"],
                    "host_dev_name": iface["host_dev_name"],
                }
                if guestMac, ok := iface["guest_mac"].(string); ok {
                    newIface["guest_mac"] = guestMac
                }
                newInterfaces = append(newInterfaces, newIface)
            }
        }
        d.Set("network_interfaces", newInterfaces)
    }

    tflog.Debug(ctx, "Firecracker VM data source read completed", map[string]interface{}{
        "id": vmID,
    })

    return diags
}
