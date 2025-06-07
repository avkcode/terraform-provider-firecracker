package firecracker

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "time"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceFirecrackerVM defines the schema and lifecycle methods for the firecracker_vm resource.
func resourceFirecrackerVM() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceFirecrackerVMCreate,
        ReadContext:   resourceFirecrackerVMRead,
        UpdateContext: resourceFirecrackerVMUpdate,
        DeleteContext: resourceFirecrackerVMDelete,
        Schema: map[string]*schema.Schema{
            "kernel_image_path": {
                Type:         schema.TypeString,
                Required:     true,
                Description:  "Path to the kernel image.",
                ValidateFunc: validation.StringIsNotEmpty,
            },
            "boot_args": {
                Type:        schema.TypeString,
                Optional:    true,
                Default:     "console=ttyS0 noapic reboot=k panic=1 pci=off root=/dev/vda rootfstype=ext4 rw init=/sbin/init",
                Description: "Boot arguments for the kernel.",
            },
            "drives": {
                Type:        schema.TypeList,
                Required:    true,
                Description: "List of drives attached to the VM.",
                MinItems:    1,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "drive_id": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "ID of the drive.",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "path_on_host": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "Path to the drive on the host.",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "is_root_device": {
                            Type:        schema.TypeBool,
                            Required:    true,
                            Description: "Whether this drive is the root device.",
                        },
                        "is_read_only": {
                            Type:        schema.TypeBool,
                            Optional:    true,
                            Default:     false,
                            Description: "Whether the drive is read-only.",
                        },
                    },
                },
            },
            "machine_config": {
                Type:        schema.TypeList,
                MaxItems:    1,
                Required:    true,
                Description: "Machine configuration for the VM.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "vcpu_count": {
                            Type:         schema.TypeInt,
                            Required:     true,
                            Description:  "Number of vCPUs.",
                            ValidateFunc: validation.IntAtLeast(1),
                        },
                        "mem_size_mib": {
                            Type:         schema.TypeInt,
                            Required:     true,
                            Description:  "Memory size in MiB.",
                            ValidateFunc: validation.IntAtLeast(128),
                        },
                    },
                },
            },
            "network_interfaces": {
                Type:        schema.TypeList,
                Optional:    true,
                Description: "List of network interfaces attached to the VM.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "iface_id": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "ID of the network interface.",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "host_dev_name": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "Host device name for the interface.",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "guest_mac": {
                            Type:         schema.TypeString,
                            Optional:     true,
                            Description:  "MAC address for the guest.",
                            ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`), "must be a valid MAC address"),
                        },
                    },
                },
            },
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(10 * time.Minute),
        },
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
    }
}

// resourceFirecrackerVMCreate creates a new Firecracker VM.
func resourceFirecrackerVMCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)
    var diags diag.Diagnostics

    // Generate a unique ID for the VM
    vmID := uuid.New().String()
    d.SetId(vmID)

    tflog.Info(ctx, "Creating Firecracker VM", map[string]interface{}{
        "id": vmID,
    })

    // Construct the boot source payload
    bootSource := map[string]interface{}{
        "kernel_image_path": d.Get("kernel_image_path").(string),
        "boot_args":         d.Get("boot_args").(string),
    }

    // Construct the drives payload
    drives := []map[string]interface{}{}
    for _, rawDrive := range d.Get("drives").([]interface{}) {
        drive := rawDrive.(map[string]interface{})
        drives = append(drives, map[string]interface{}{
            "drive_id":       drive["drive_id"].(string),
            "path_on_host":   drive["path_on_host"].(string),
            "is_root_device": drive["is_root_device"].(bool),
            "is_read_only":   drive["is_read_only"].(bool),
        })
    }

    // Construct the machine config payload
    machineConfigRaw := d.Get("machine_config").([]interface{})[0].(map[string]interface{})
    machineConfig := map[string]interface{}{
        "vcpu_count":   machineConfigRaw["vcpu_count"].(int),
        "mem_size_mib": machineConfigRaw["mem_size_mib"].(int),
    }

    // Construct the network interfaces payload
    networkInterfaces := []map[string]interface{}{}
    for _, rawIface := range d.Get("network_interfaces").([]interface{}) {
        iface := rawIface.(map[string]interface{})
        ifaceMap := map[string]interface{}{
            "iface_id":      iface["iface_id"].(string),
            "host_dev_name": iface["host_dev_name"].(string),
        }
        
        // Only add guest_mac if it's set
        if mac, ok := iface["guest_mac"].(string); ok && mac != "" {
            ifaceMap["guest_mac"] = mac
        }
        
        networkInterfaces = append(networkInterfaces, ifaceMap)
    }

    // Construct the full payload
    payload := map[string]interface{}{
        "boot-source":        bootSource,
        "drives":             drives,
        "machine-config":     machineConfig,
        "network-interfaces": networkInterfaces,
        "vm-id":              vmID,
    }

    // Send the request to the Firecracker API
    err := client.CreateVM(ctx, payload)
    if err != nil {
        return diag.FromErr(fmt.Errorf("failed to create VM: %w", err))
    }

    tflog.Info(ctx, "Firecracker VM created successfully", map[string]interface{}{
        "id": vmID,
    })

    // Read the resource to ensure state is consistent
    return resourceFirecrackerVMRead(ctx, d, m)
}

func resourceFirecrackerVMRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)
    var diags diag.Diagnostics

    vmID := d.Id()
    tflog.Debug(ctx, "Reading Firecracker VM", map[string]interface{}{
        "id": vmID,
    })

    // Get VM details from the API
    vmInfo, err := client.GetVM(ctx, vmID)
    if err != nil {
        return diag.FromErr(fmt.Errorf("error reading VM: %w", err))
    }

    // If VM not found, remove from state
    if vmInfo == nil {
        tflog.Warn(ctx, "Firecracker VM not found, removing from state", map[string]interface{}{
            "id": vmID,
        })
        d.SetId("")
        return diags
    }

    // Update the resource data based on the VM info
    // This is a simplified example - you would need to adapt this to match
    // the actual structure of your API response
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

    tflog.Debug(ctx, "Firecracker VM read completed", map[string]interface{}{
        "id": vmID,
    })

    return diags
}

func resourceFirecrackerVMUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)
    vmID := d.Id()
    
    tflog.Info(ctx, "Updating Firecracker VM", map[string]interface{}{
        "id": vmID,
    })
    
    // Check which fields have changed
    var updatePayload = make(map[string]interface{})
    
    // Handle machine config updates
    if d.HasChange("machine_config") {
        machineConfigRaw := d.Get("machine_config").([]interface{})[0].(map[string]interface{})
        updatePayload["machine-config"] = map[string]interface{}{
            "vcpu_count":   machineConfigRaw["vcpu_count"].(int),
            "mem_size_mib": machineConfigRaw["mem_size_mib"].(int),
        }
    }
    
    // Handle network interface updates
    if d.HasChange("network_interfaces") {
        networkInterfaces := []map[string]interface{}{}
        for _, rawIface := range d.Get("network_interfaces").([]interface{}) {
            iface := rawIface.(map[string]interface{})
            ifaceMap := map[string]interface{}{
                "iface_id":      iface["iface_id"].(string),
                "host_dev_name": iface["host_dev_name"].(string),
            }
            
            if mac, ok := iface["guest_mac"].(string); ok && mac != "" {
                ifaceMap["guest_mac"] = mac
            }
            
            networkInterfaces = append(networkInterfaces, ifaceMap)
        }
        updatePayload["network-interfaces"] = networkInterfaces
    }
    
    // If there are changes to apply
    if len(updatePayload) > 0 {
        err := client.UpdateVM(ctx, vmID, updatePayload)
        if err != nil {
            return diag.FromErr(fmt.Errorf("failed to update VM: %w", err))
        }
        
        tflog.Info(ctx, "Firecracker VM updated successfully", map[string]interface{}{
            "id": vmID,
        })
    } else {
        tflog.Debug(ctx, "No changes to apply for Firecracker VM", map[string]interface{}{
            "id": vmID,
        })
    }
    
    // Read the resource to ensure state is consistent
    return resourceFirecrackerVMRead(ctx, d, m)
}

func resourceFirecrackerVMDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)
    var diags diag.Diagnostics
    
    vmID := d.Id()
    tflog.Info(ctx, "Deleting Firecracker VM", map[string]interface{}{
        "id": vmID,
    })
    
    err := client.DeleteVM(ctx, vmID)
    if err != nil {
        return diag.FromErr(fmt.Errorf("error deleting VM: %w", err))
    }
    
    // Remove the VM from state
    d.SetId("")
    
    tflog.Info(ctx, "Firecracker VM deleted successfully")
    
    return diags
}
