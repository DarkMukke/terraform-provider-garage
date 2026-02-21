provider "garage" {
  endpoints = {
    admin = "http://localhost:3903" # Admin API
    s3    = "http://localhost:3900" # S3 API
  }
  token      = "admin-token"
  access_key = "GK123..."     # S3 access key
  secret_key = "secret123..." # S3 secret key
}