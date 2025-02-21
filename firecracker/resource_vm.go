package firecracker

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
                Type:        schema.TypeString,
                Required:    true,
                Description: "Path to the kernel image.",
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
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "drive_id": {
                            Type:        schema.TypeString,
                            Required:    true,
                            Description: "ID of the drive.",
                        },
                        "path_on_host": {
                            Type:        schema.TypeString,
                            Required:    true,
                            Description: "Path to the drive on the host.",
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
                            Type:        schema.TypeInt,
                            Required:    true,
                            Description: "Number of vCPUs.",
                        },
                        "mem_size_mib": {
                            Type:        schema.TypeInt,
                            Required:    true,
                            Description: "Memory size in MiB.",
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
                            Type:        schema.TypeString,
                            Required:    true,
                            Description: "ID of the network interface.",
                        },
                        "host_dev_name": {
                            Type:        schema.TypeString,
                            Required:    true,
                            Description: "Host device name for the interface.",
                        },
                        "guest_mac": {
                            Type:        schema.TypeString,
                            Optional:    true,
                            Description: "MAC address for the guest.",
                        },
                    },
                },
            },
        },
    }
}

// resourceFirecrackerVMCreate creates a new Firecracker VM.
func resourceFirecrackerVMCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    client := m.(*FirecrackerClient)

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
        networkInterfaces = append(networkInterfaces, map[string]interface{}{
            "iface_id":      iface["iface_id"].(string),
            "host_dev_name": iface["host_dev_name"].(string),
            "guest_mac":     iface["guest_mac"].(string),
        })
    }

    // Construct the full payload
    payload := map[string]interface{}{
        "boot-source":         bootSource,
        "drives":              drives,
        "machine-config":      machineConfig,
        "network-interfaces":  networkInterfaces,
    }

    // Convert the payload to JSON
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return diag.FromErr(fmt.Errorf("failed to marshal payload: %w", err))
    }

    // Send the request to the Firecracker API
    url := client.BaseURL + "/vm"
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
    if err != nil {
        return diag.FromErr(fmt.Errorf("failed to send request: %w", err))
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent {
        return diag.FromErr(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
    }

    // Set the resource ID
    d.SetId(fmt.Sprintf("%s-%s", d.Get("kernel_image_path"), d.Get("drives").([]interface{})[0].(map[string]interface{})["path_on_host"]))

    return nil
}

func resourceFirecrackerVMRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    // Implement read logic if needed
    return nil
}

func resourceFirecrackerVMUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    // Implement update logic if needed
    return nil
}

func resourceFirecrackerVMDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    // Implement delete logic if needed
    return nil
}
