terraform{
    required_providers {
        kafka = {
            source  = "Mongey/kafka"
        }
        aws = {
            source  = "hashicorp/aws"
            version = "~> 6.0"
        }
    }
}

provider "kafka" {
    bootstrap_servers = ["kafka:9092"]
    tls_enabled = false
}

provider "aws" {
  region     = "us-east-1"
  access_key = var.s3_access_key
  secret_key = var.s3_secret_key

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true

  endpoints {
    s3 = "http://localhost:8333"
  }
}