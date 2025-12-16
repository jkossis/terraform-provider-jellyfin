terraform {
  required_providers {
    jellyfin = {
      source = "registry.terraform.io/jkossis/jellyfin"
    }
  }
}

provider "jellyfin" {
  # Configuration can be set here or via environment variables:
  # JELLYFIN_ENDPOINT - The Jellyfin server URL
  # JELLYFIN_USERNAME - The username for authentication
  # JELLYFIN_PASSWORD - The password for authentication

  # endpoint = "http://localhost:8096"
  # username = "admin"
  # password = "your-password-here"
}
