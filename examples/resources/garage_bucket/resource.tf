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


# Basic bucket
resource "garage_bucket" "example" {
  global_alias = "my-bucket"
}

# Bucket with website hosting
resource "garage_bucket" "website" {
  global_alias           = "my-website"
  website_enabled        = true
  website_index_document = "index.html"
  website_error_document = "error.html"
}

# Bucket with quotas
resource "garage_bucket" "limited" {
  global_alias = "limited-bucket"
  max_size     = 1073741824 # 1 GB in bytes
  max_objects  = 10000
}

# Bucket with all options
resource "garage_bucket" "full" {
  global_alias           = "full-featured-bucket"
  website_enabled        = true
  website_index_document = "index.html"
  website_error_document = "404.html"
  max_size               = 10737418240 # 10 GB in bytes
  max_objects            = 100000
}
