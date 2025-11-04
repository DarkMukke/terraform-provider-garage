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

# Alternatively, use environment variables:
# export GARAGE_ENDPOINT="http://localhost:3903"
# export GARAGE_TOKEN="your-admin-token-here"
