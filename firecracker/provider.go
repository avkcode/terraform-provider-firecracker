package firecracker

import (
    "context"
 
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FirecrackerClient represents the client for interacting with the Firecracker API.
type FirecrackerClient struct {
    BaseURL string
}

// Provider defines the Terraform provider for Firecracker.
func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "base_url": {
                Type:        schema.TypeString,
                Required:    true,
                Description: "The base URL for the Firecracker API.",
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "firecracker_vm": resourceFirecrackerVM(),
        },
        ConfigureContextFunc: configureProvider,
    }
}

// configureProvider initializes the FirecrackerClient with the provided configuration.
func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
    baseURL := d.Get("base_url").(string)
    return &FirecrackerClient{
        BaseURL: baseURL,
    }, nil
}
