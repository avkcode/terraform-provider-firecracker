package firecracker

import (
    "context"
    "net/http"
    "time"
 
    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FirecrackerClient represents the client for interacting with the Firecracker API.
// It handles communication with the Firecracker HTTP API for managing microVMs.
type FirecrackerClient struct {
    BaseURL    string
    HTTPClient httpClient
    Timeout    time.Duration
}

// Provider returns a *schema.Provider for Firecracker.
// It defines the provider's configuration, resources, and data sources.
func Provider() *schema.Provider {
    p := &schema.Provider{
        Schema: map[string]*schema.Schema{
            "base_url": {
                Type:        schema.TypeString,
                Required:    true,
                Description: "The base URL for the Firecracker API.",
            },
            "timeout": {
                Type:        schema.TypeInt,
                Optional:    true,
                Default:     30,
                Description: "Timeout in seconds for API operations.",
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "firecracker_vm": resourceFirecrackerVM(),
        },
        DataSourcesMap: map[string]*schema.Resource{
            "firecracker_vm": dataSourceFirecrackerVM(),
        },
        ConfigureContextFunc: configureProvider,
    }
    
    return p
}

// configureProvider initializes the FirecrackerClient with the provided configuration.
// It creates an HTTP client with appropriate timeouts and connection settings.
func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
    baseURL := d.Get("base_url").(string)
    timeout := d.Get("timeout").(int)
    
    tflog.Info(ctx, "Configuring Firecracker provider", map[string]interface{}{
        "base_url": baseURL,
        "timeout":  timeout,
    })
    
    httpClient := &http.Client{
        Timeout: time.Duration(timeout) * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 20,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    return &FirecrackerClient{
        BaseURL:    baseURL,
        HTTPClient: httpClient,
        Timeout:    time.Duration(timeout) * time.Second,
    }, nil
}
