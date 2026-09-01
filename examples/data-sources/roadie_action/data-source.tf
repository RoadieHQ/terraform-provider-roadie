data "roadie_action" "existing" {
  slug = "list-repositories"
}

output "action_enabled" {
  value = data.roadie_action.existing.enabled
}
