terraform {
  required_providers {
    roadie = {
      source = "RoadieHQ/roadie"
    }
  }
}

provider "roadie" {
  host         = "https://app-api.roadie.so"
  api_token    = var.roadie_api_token
  workspace_id = var.roadie_workspace_id
}

variable "roadie_api_token" {
  type      = string
  sensitive = true
}

variable "roadie_workspace_id" {
  type = string
}
