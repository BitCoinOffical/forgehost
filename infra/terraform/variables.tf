variable "s3_access_key" {
  type        = string
  description = "Access key for S3 compatible storage"
  sensitive   = true
}

variable "s3_secret_key" {
  type        = string
  description = "Secret key for S3 compatible storage"
  sensitive   = true
}