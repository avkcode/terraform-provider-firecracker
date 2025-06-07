package firecracker

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceFirecrackerVM defines the schema and CRUD operations for the firecracker_vm resource.
// This resource allows users to create, read, update, and delete Firecracker microVMs.
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
                Description:  "Path to the kernel image. Must be accessible by the Firecracker process. This should be an uncompressed Linux kernel binary (vmlinux format).",
                ValidateFunc: validation.StringIsNotEmpty,
            },
            "boot_args": {
                Type:        schema.TypeString,
                Optional:    true,
                Default:     "console=ttyS0 noapic reboot=k panic=1 pci=off root=/dev/vda rootfstype=ext4 rw init=/sbin/init",
                Description: "Boot arguments for the kernel. These are passed to the kernel at boot time. The default arguments are suitable for most Linux distributions with an ext4 root filesystem.",
            },
            "drives": {
                Type:        schema.TypeList,
                Required:    true,
                Description: "List of drives attached to the VM. At least one drive must be specified, typically containing the root filesystem.",
                MinItems:    1,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "drive_id": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "ID of the drive. This is used to identify the drive within Firecracker and must be unique within the VM.",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "path_on_host": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "Path to the drive on the host. This must be accessible by the Firecracker process and should be a valid disk image (e.g., ext4 filesystem).",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "is_root_device": {
                            Type:        schema.TypeBool,
                            Required:    true,
                            Description: "Whether this drive is the root device. Only one drive can be marked as the root device. This should be set to true for the drive containing the root filesystem.",
                        },
                        "is_read_only": {
                            Type:        schema.TypeBool,
                            Optional:    true,
                            Default:     false,
                            Description: "Whether the drive is read-only. Set to true for immutable drives like OS images, and false for drives that need to persist data.",
                        },
                    },
                },
            },
            "machine_config": {
                Type:        schema.TypeList,
                MaxItems:    1,
                Required:    true,
                Description: "Machine configuration for the VM. This defines the virtual hardware resources allocated to the VM.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "vcpu_count": {
                            Type:         schema.TypeInt,
                            Required:     true,
                            Description:  "Number of vCPUs. Must be between 1 and 32.",
                            ValidateFunc: validation.IntBetween(1, 32),
                        },
                        "mem_size_mib": {
                            Type:         schema.TypeInt,
                            Required:     true,
                            Description:  "Memory size in MiB. Must be between 128 and 32768.",
                            ValidateFunc: validation.IntBetween(128, 32768),
                        },
                    },
                },
            },
            "network_interfaces": {
                Type:        schema.TypeList,
                Optional:    true,
                Description: "List of network interfaces attached to the VM. Each interface connects to a TAP device on the host.",
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "iface_id": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "ID of the network interface. This is used to identify the interface within Firecracker and must be unique within the VM.",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "host_dev_name": {
                            Type:         schema.TypeString,
                            Required:     true,
                            Description:  "Host device name for the interface. This should be a TAP device that exists on the host (e.g., 'tap0').",
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "guest_mac": {
                            Type:         schema.TypeString,
                            Optional:     true,
                            Description:  "MAC address for the guest network interface. If not specified, Firecracker will generate one. Format: 'XX:XX:XX:XX:XX:XX'.",
                            ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`), "must be a valid MAC address"),
                        },
                    },
                },
            },
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(5 * time.Minute),
            Delete: schema.DefaultTimeout(5 * time.Minute),
            Read:   schema.DefaultTimeout(1 * time.Minute),
        },
        Importer: &schema.ResourceImporter{
            StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
                client := meta.(*FirecrackerClient)
                vmID := d.Id()
                
                tflog.Info(ctx, "Importing Firecracker VM", map[string]interface{}{
                    "id": vmID,
                })
                
                // Get VM details from API
                vmInfo, err := client.GetVM(ctx, vmID)
                if err != nil {
                    return nil, fmt.Errorf("error importing VM %s: %w", vmID, err)
                }
                
                if vmInfo == nil {
                    return nil, fmt.Errorf("VM with ID %s not found", vmID)
                }
                
                // Read the resource data from the imported VM
                d.SetId(vmID)
                resourceFirecrackerVMRead(ctx, d, meta)
                
                return []*schema.ResourceData{d}, nil
            },
        },
    }
}

// resourceFirecrackerVMCreate creates a new Firecracker VM.
func resourceFirecrackerVMCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)

    // Generate a unique ID for the VM
    vmID := uuid.New().String()
    d.SetId(vmID)

    tflog.Info(ctx, "Creating Firecracker VM", map[string]interface{}{
        "id": vmID,
    })

    // Get boot args and ensure it has the correct root device specification
    bootArgs := d.Get("boot_args").(string)
    
    // Check if boot_args already contains root=PARTUUID=
    if !strings.Contains(bootArgs, "root=PARTUUID=") {
        // If it contains root=/dev/vda, replace it with root=PARTUUID=rootfs
        if strings.Contains(bootArgs, "root=/dev/vda") {
            bootArgs = strings.Replace(bootArgs, "root=/dev/vda", "root=PARTUUID=rootfs", 1)
        } else {
            // If it doesn't contain either, append root=PARTUUID=rootfs
            bootArgs += " root=PARTUUID=rootfs"
        }
    }
    
    // Construct the boot source payload
    bootSource := map[string]interface{}{
        "kernel_image_path": d.Get("kernel_image_path").(string),
        "boot_args":         bootArgs,
    }

    // Construct the drives payload
    drives := []map[string]interface{}{}
    for _, rawDrive := range d.Get("drives").([]interface{}) {
        drive := rawDrive.(map[string]interface{})
        driveMap := map[string]interface{}{
            "drive_id":       drive["drive_id"].(string),
            "path_on_host":   drive["path_on_host"].(string),
            "is_root_device": drive["is_root_device"].(bool),
            "is_read_only":   drive["is_read_only"].(bool),
        }
        
        // Explicitly convert to bool to ensure proper type for Firecracker API
        isRootDevice, ok := drive["is_root_device"].(bool)
        if !ok {
            if strVal, ok := drive["is_root_device"].(string); ok {
                isRootDevice = strVal == "true"
            }
        }
        driveMap["is_root_device"] = isRootDevice
        
        isReadOnly, ok := drive["is_read_only"].(bool)
        if !ok {
            if strVal, ok := drive["is_read_only"].(string); ok {
                isReadOnly = strVal == "true"
            }
        }
        driveMap["is_read_only"] = isReadOnly
        
        // Log the drive configuration for debugging
        tflog.Debug(ctx, "Drive configuration", map[string]interface{}{
            "drive_id":       driveMap["drive_id"],
            "path_on_host":   driveMap["path_on_host"],
            "is_root_device": driveMap["is_root_device"],
            "is_read_only":   driveMap["is_read_only"],
        })
        
        // Log drive configuration for debugging
        tflog.Debug(ctx, "Configuring drive for VM", map[string]interface{}{
            "drive_id":       driveMap["drive_id"],
            "path_on_host":   driveMap["path_on_host"],
            "is_root_device": driveMap["is_root_device"],
            "is_read_only":   driveMap["is_read_only"],
        })
        
        drives = append(drives, driveMap)
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
    
    // Set the ID to ensure it's properly tracked in state
    d.SetId(vmID)

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
    var hasChanges bool
    
    // Log changes that would require VM recreation
    if d.HasChange("machine_config") {
        tflog.Warn(ctx, "Machine configuration changes require VM recreation", map[string]interface{}{
            "id": vmID,
        })
        hasChanges = true
    }
    
    if d.HasChange("network_interfaces") {
        tflog.Warn(ctx, "Network interface changes require VM recreation", map[string]interface{}{
            "id": vmID,
        })
        hasChanges = true
    }
    
    if d.HasChange("kernel_image_path") || d.HasChange("boot_args") {
        tflog.Warn(ctx, "Boot configuration changes require VM recreation", map[string]interface{}{
            "id": vmID,
        })
        hasChanges = true
    }
    
    if d.HasChange("drives") {
        tflog.Warn(ctx, "Drive configuration changes require VM recreation", map[string]interface{}{
            "id": vmID,
        })
        hasChanges = true
    }
    
    // If there are changes, call the API (which will just log a warning)
    if hasChanges {
        err := client.UpdateVM(ctx, vmID, nil)
        if err != nil {
            return diag.FromErr(fmt.Errorf("failed to update VM: %w", err))
        }
        
        tflog.Info(ctx, "Firecracker VM update processed (note: most changes require recreation)", map[string]interface{}{
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
