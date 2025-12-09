terraform {
  required_providers {
    garage = {
      source = "darkmukke/garage"
    }
  }
}

provider "garage" {
  endpoints = {
    admin = "http://localhost:3903"  # Admin API
    s3    = "http://localhost:3900"  # S3 API
  }
  token      = "admin-token"
  access_key = "GK123..."  # S3 access key
  secret_key = "secret123..." # S3 secret key
}

# Look up bucket by global alias
data "garage_bucket" "by_alias" {
  global_alias = "my-bucket"
}

# Look up bucket by ID
data "garage_bucket" "by_id" {
  id = "8d7c3c6e-7b9d-4c3a-9f2e-1a5b6c7d8e9f"
}

# Use data source output
output "bucket_info" {
  value = {
    id                 = data.garage_bucket.by_alias.id
    aliases            = data.garage_bucket.by_alias.global_aliases
    website_enabled    = data.garage_bucket.by_alias.website_enabled
    objects            = data.garage_bucket.by_alias.objects
    bytes              = data.garage_bucket.by_alias.bytes
    unfinished_uploads = data.garage_bucket.by_alias.unfinished_uploads
  }
}
