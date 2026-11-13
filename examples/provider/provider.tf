terraform {
  required_providers {
    opnsense = {
      source  = "dev.io/inkstom/opnsense"
      version = "1.0.0"
    }
  }
}

provider "opnsense" {
  uri        = "https://opnsense.example.com"
  api_key    = "..."
  api_secret = "..."
}
