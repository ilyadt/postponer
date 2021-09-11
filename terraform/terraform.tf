terraform {
  backend "s3" {
    bucket = "terraform-postponer"
    key    = "state.tfstate"
  }
}
