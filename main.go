package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
    "github.com/avkcode/terraform-provider-firecracker/firecracker"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        ProviderFunc: firecracker.Provider,
    })
}
