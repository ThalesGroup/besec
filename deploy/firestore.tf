resource "google_storage_bucket" "firestore-backup" {
  name                        = "firestore-backup-${var.gcp-project}"
  storage_class               = "NEARLINE"
  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = "365"
    }
    action {
      type = "Delete"
    }
  }
}

// service account used for backup (e.g. default compute account project@appspot.gserviceaccount.com) needs datastore.databases.export
// and roles/storage.admin on the bucket
// it would be better to allocate only objectWriter to the pipeline role, but it already has project-wide storage.admin.
