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


# Auto-generated access key with a name
resource "garage_key" "example" {
  name = "my-application-key"
}

# Auto-generated access key without a name
resource "garage_key" "unnamed" {
}

# Import an existing key with predefined credentials
resource "garage_key" "imported" {
  id                = "GK31c2f218a2e44f485b94239e"
  secret_access_key = "7d37d093435a75809f8f090b072de87928d1c355db9d9340431b28e776374705"
  name              = "imported-key"
}

# Import a key without a name
resource "garage_key" "imported_no_name" {
  id                = "GKaf8f07b8e6f44c5a9c7b3d2e"
  secret_access_key = "3c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d"
}

# Output the credentials
output "access_key_id" {
  value = garage_key.example.id
}

output "secret_access_key" {
  value     = garage_key.example.secret_access_key
  sensitive = true
}

output "imported_key_id" {
  value = garage_key.imported.id
}
