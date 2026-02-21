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


resource "garage_object" "example" {
  bucket       = garage_bucket.example.id
  key          = "hello.txt"
  content      = "Hello, Garage!"
  content_type = "text/plain"
}

# Or with a file
resource "garage_object" "file_example" {
  bucket = garage_bucket.example.id
  key    = "data.json"
  source = "${path.module}/data.json"
}
