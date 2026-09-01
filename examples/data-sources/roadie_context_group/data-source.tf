data "roadie_context_group" "existing" {
  slug = "services-and-teams"
}

output "context_group_name" {
  value = data.roadie_context_group.existing.name
}
