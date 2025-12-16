# Look up an API key by application name
data "jellyfin_api_key" "by_name" {
  app_name = "My Terraform Application"
}

# Look up an API key by access token
data "jellyfin_api_key" "by_token" {
  access_token = "your-api-key-token-here"
}

output "found_key_created_at" {
  value = data.jellyfin_api_key.by_name.date_created
}
