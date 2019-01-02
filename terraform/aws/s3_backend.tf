terraform {
  backend "s3" {
    bucket                      = "terraform-machine-controller"
    endpoint                    = "http://minio.minio:9000"
    access_key                  = "PMIC1HMXNB2R67RNPIX8"
    secret_key                  = "NemiWx+uY79rcJ0hXrktzHk1dm9c0k85WepbuSlK"
    region                      = "myregion"
    skip_region_validation      = "true"
    skip_metadata_api_check     = "true"
    skip_requesting_account_id  = "true"
    skip_credentials_validation = "true"
    force_path_style            = "true"
  }
}
