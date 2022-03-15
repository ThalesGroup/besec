resource "google_storage_bucket" "image-store" {
  name               = "artifacts.${var.gcp-project}.appspot.com"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_binding" "public_read" {
  bucket = google_storage_bucket.image-store.name
  role   = "roles/storage.objectViewer"
  members = [
    "allUsers",
  ]
}

