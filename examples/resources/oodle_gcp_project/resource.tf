# One oodle_gcp_project resource is one GCP project. To collect several
# projects, declare one resource per project, or use for_each as below.
#
# Oodle never holds a key. The customer service account grants
# oodle_principal the Service Account Token Creator role, and Oodle
# impersonates it to read Cloud Monitoring. Without that grant the
# integration stays in NOT_CONNECTED and no metrics arrive.

# Minimal: one project, with the service account already created.
resource "oodle_gcp_project" "prod" {
  project                  = "acme-prod"
  customer_service_account = "oodle@acme-prod.iam.gserviceaccount.com"
}

# Several projects, each addressable on its own. A change to one project
# leaves the others untouched, which is what a shared list of projects
# could not do.
locals {
  gcp_projects = {
    prod    = "acme-prod"
    staging = "acme-staging"
  }
}

resource "oodle_gcp_project" "collected" {
  for_each = local.gcp_projects

  project                  = each.value
  customer_service_account = "oodle@${each.value}.iam.gserviceaccount.com"
}

# End to end: create the service account, grant it the viewer roles Oodle
# reads with, register the project, then let Oodle impersonate the account.
resource "google_service_account" "oodle" {
  project      = "acme-prod"
  account_id   = "oodle"
  display_name = "Oodle metric collection"
}

resource "google_project_iam_member" "oodle" {
  for_each = toset([
    "roles/browser",
    "roles/compute.viewer",
    "roles/monitoring.viewer",
    "roles/cloudasset.viewer",
  ])

  project = "acme-prod"
  role    = each.value
  member  = "serviceAccount:${google_service_account.oodle.email}"
}

resource "oodle_gcp_project" "managed" {
  project                  = "acme-prod"
  customer_service_account = google_service_account.oodle.email

  depends_on = [google_project_iam_member.oodle]
}

# The grant that replaces a service account key. oodle_principal is read
# back from the integration, so the address is never written out here and
# stays right if Oodle ever changes it.
resource "google_service_account_iam_member" "oodle_token_creator" {
  service_account_id = google_service_account.oodle.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${oodle_gcp_project.managed.oodle_principal}"
}
