terraform {
  required_providers {
    garage = {
      source = "darkmukke/garage"
    }
  }
}

provider "garage" {
  endpoints = {
    admin = "http://localhost:3903" # Admin API
    s3    = "http://localhost:3900" # S3 API
  }
  token      = "admin-token"
  access_key = "GK123..."     # S3 access key
  secret_key = "secret123..." # S3 secret key
}


# Read a text file from Garage
data "garage_object" "config" {
  bucket = "my-bucket"
  key    = "config.json"
}

# Use the downloaded content
output "config_content" {
  value = jsondecode(data.garage_object.config.body)
}

# Check file metadata
output "file_info" {
  value = {
    size         = data.garage_object.config.content_length
    content_type = data.garage_object.config.content_type
    etag         = data.garage_object.config.etag
    modified     = data.garage_object.config.last_modified
  }
}

# Use with local file output
resource "local_file" "downloaded" {
  content  = data.garage_object.config.body
  filename = "${path.module}/downloaded-config.json"
}

data "garage_object" "image" {
  bucket = "media-bucket"
  key    = "logo.png"
}

# Write binary file (requires local_file resource with content_base64)
resource "local_file" "image" {
  content_base64 = base64encode(data.garage_object.image.body)
  filename       = "${path.module}/logo.png"
}
