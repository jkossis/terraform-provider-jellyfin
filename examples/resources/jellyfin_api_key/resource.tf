# Create a new API key for an application
resource "jellyfin_api_key" "example" {
  app_name = "My Terraform Application"
}

# Output the generated access token (sensitive)
output "api_key_token" {
  value     = jellyfin_api_key.example.access_token
  sensitive = true
}

output "api_key_created_at" {
  value = jellyfin_api_key.example.date_created
}
