# Terraform Provider for Jellyfin

A Terraform provider for managing [Jellyfin](https://jellyfin.org/) media server resources.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Installation

The provider can be installed from the Terraform Registry:

```hcl
terraform {
  required_providers {
    jellyfin = {
      source = "jkossis/jellyfin"
    }
  }
}
```

## Usage

```hcl
provider "jellyfin" {
  endpoint = "http://localhost:8096"
  api_key  = "your-api-key-here"
}

# Create an API key
resource "jellyfin_api_key" "example" {
  app_name = "My Application"
}

# Look up an existing API key
data "jellyfin_api_key" "existing" {
  app_name = "Existing Application"
}
```

### Authentication

The provider requires a Jellyfin API key for authentication. You can generate one from the Jellyfin dashboard under **Administration > API Keys**.

Configuration can be provided via:
- Provider configuration block (`endpoint` and `api_key`)
- Environment variables (`JELLYFIN_ENDPOINT` and `JELLYFIN_API_KEY`)

## Resources

- `jellyfin_api_key` - Manages Jellyfin API keys

## Data Sources

- `jellyfin_api_key` - Retrieves information about existing API keys

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider:

```shell
go build -o terraform-provider-jellyfin
```

## Developing the Provider

To generate or update documentation, run `make generate`.

To run the full suite of acceptance tests:

```shell
make testacc
```

*Note:* Acceptance tests require a running Jellyfin instance and valid credentials.

## License

MPL-2.0
