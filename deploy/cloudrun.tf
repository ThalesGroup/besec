##
## CLOUD RUN
##

variable "image-tag" {
  type        = string
  description = "The tag to use for the container image"
}
variable "region" {
  type        = string
  description = "The Google Cloud region to deploy the Cloud Run service"
}
variable "slack-webhook-name" {
  type        = string
  description = "The name of a configured slack webhook to use for alerts"
}
variable "domain" {
  type        = string
  description = "The domain name the Cloud Run service is available at"
}

resource "google_cloud_run_service" "prod" {
  name     = "besec"
  location = var.region
  provider = google

  metadata {
    namespace = var.gcp-project
  }

  template {
    spec {
      containers {
        image = "gcr.io/${var.gcp-project}/besec:${var.image-tag}"
        args = ["serve",
          "--gcp-project=${var.gcp-project}",
          "--slack-webhook-name=${var.slack-webhook-name}",
          "--stackdriver-logs",
          "-v",
          "--no-emulator",
        ]
      }
      service_account_name = google_service_account.cloud-run.email
    }
  }
}

data "google_iam_policy" "public" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "public" {
  location = google_cloud_run_service.prod.location
  project  = google_cloud_run_service.prod.project
  service  = google_cloud_run_service.prod.name

  policy_data = data.google_iam_policy.public.policy_data
}

resource "google_cloud_run_domain_mapping" "prod" {
  location = var.region
  provider = google
  name     = var.domain

  metadata {
    namespace = var.gcp-project
  }

  spec {
    route_name = google_cloud_run_service.prod.name
  }
}

resource "google_service_account" "cloud-run" {
  account_id   = "cloud-run"
  display_name = "Cloud Run service account"
  description  = "The runtime service account for BeSec application"
}
resource "google_project_iam_member" "cloud-run-datastore-perms" {
  project = var.gcp-project
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.cloud-run.email}"
}

# needed to deploy the service https://cloud.google.com/run/docs/securing/service-identity?hl=en#permissions_required_to_use_non-default_identities
resource "google_service_account_iam_member" "pipeline-assume-cloudrun-svcacct" {
  service_account_id = google_service_account.cloud-run.name
  role               = "roles/iam.serviceAccountUser"

  member = "serviceAccount:${google_service_account.pipeline.email}"
}