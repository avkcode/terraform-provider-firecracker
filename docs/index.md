# Firecracker Provider

The Firecracker provider allows Terraform to manage [Firecracker](https://firecracker-microvm.github.io/) microVMs. Firecracker is a virtualization technology that enables you to implement lightweight virtual machines (microVMs) with minimal overhead.

## Example Usage

```hcl
terraform {
  required_providers {
    firecracker = {
      source = "avkcode/firecracker"
      version = "0.1.0"
    }
  }
}

provider "firecracker" {
  base_url = "http://localhost:8080"
  timeout  = 30
}
```

## Authentication

The Firecracker provider communicates with the Firecracker API over HTTP. No authentication is required by default, but you should ensure that the API socket is properly secured.

## Provider Arguments

* `base_url` - (Required) The base URL of the Firecracker API. This is typically a local URL like `http://localhost:8080` that forwards to the Firecracker API socket.
* `timeout` - (Optional) Timeout in seconds for API operations. Default is 30 seconds.
