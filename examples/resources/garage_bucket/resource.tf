terraform {
  required_providers {
    garage = {
      source = "jkossis/garage"
    }
  }
}

provider "garage" {
  endpoint = "http://localhost:3903"
  token    = "your-admin-token-here"
}

# Basic bucket
resource "garage_bucket" "example" {
  global_alias = "my-bucket"
}

# Bucket with website hosting
resource "garage_bucket" "website" {
  global_alias             = "my-website"
  website_enabled          = true
  website_index_document   = "index.html"
  website_error_document   = "error.html"
}

# Bucket with quotas
resource "garage_bucket" "limited" {
  global_alias = "limited-bucket"
  max_size     = 1073741824  # 1 GB in bytes
  max_objects  = 10000
}

# Bucket with all options
resource "garage_bucket" "full" {
  global_alias             = "full-featured-bucket"
  website_enabled          = true
  website_index_document   = "index.html"
  website_error_document   = "404.html"
  max_size                 = 10737418240  # 10 GB in bytes
  max_objects              = 100000
}
