provider "google" {
  region  = var.gcp-region
  project = var.gcp-project
}

variable "gcp-project" {
  type        = string
  description = "The GCP project ID"
}
