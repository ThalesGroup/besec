variable "cli-admins" {
  type        = list
  description = "The email addresses of GCP IAM users who can use the BeSec CLI"
  default     = ["user:jane@example.com"]
}

resource "google_service_account" "cli-administrator" {
  account_id   = "cli-administrator"
  display_name = "CLI Administrator service account"
  description  = "Assumed by the CLI by administrators with permission to impersonate this account"
}

resource "google_service_account_iam_binding" "assume-cli-admin" {
  service_account_id = google_service_account.cli-administrator.name
  role               = "roles/iam.serviceAccountTokenCreator"

  members = var.cli-admins
}

resource "google_project_iam_member" "cliadmin-auth" {
  project = var.gcp-project
  role    = "roles/firebaseauth.admin"
  member  = "serviceAccount:${google_service_account.cli-administrator.email}"
}

resource "google_project_iam_member" "cliadmin-datastore" {
  project = var.gcp-project
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.cli-administrator.email}"
}

resource "google_service_account" "terraform-read" {
  account_id   = "terraform-read"
  display_name = "Terraform read only service account"
  description  = "Allows terraform plan to run in CICD unprotected branches"
}
resource "google_project_iam_member" "terraform-read-viewer" {
  project = var.gcp-project
  role    = "roles/viewer"
  member  = "serviceAccount:${google_service_account.terraform-read.email}"
}
resource "google_project_iam_member" "terraform-read-securityreviewer" {
  project = var.gcp-project
  role    = "roles/iam.securityReviewer"
  member  = "serviceAccount:${google_service_account.terraform-read.email}"
}

resource "google_service_account" "pipeline" {
  account_id   = "pipeline"
  display_name = "Pipeline service account"
  description  = "High privilege account for deploying images and applying terraform definitions in CICD"
}

variable "extra-gcp-admins" {
  type        = list
  description = "The users who need direct access to administer all GCP resources (not just management actions via the CLI)"
  default     = []
}

locals {
  admin-roles = toset(["appengine.appAdmin", "datastore.owner", "run.admin", "storage.admin"])
  gcp-admins  = toset(concat(var.extra-gcp-admins, ["serviceAccount:${google_service_account.pipeline.email}"]))
  admin-role-pairs = [
    for pair in setproduct(local.admin-roles, local.gcp-admins) : {
      role   = format("roles/%s", pair[0])
      member = pair[1]
    }
  ]
}

resource "google_project_iam_member" "admin-perms" {
  project = var.gcp-project
  # admin-role-pairs is a list, it needs to be a map to use with for_each. We don't care about the key names as long as they're unique.
  for_each = {
    for pair in local.admin-role-pairs : "${pair.role}<=>${pair.member}" => pair
  }
  role   = each.value.role
  member = each.value.member
}