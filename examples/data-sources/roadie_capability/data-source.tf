data "roadie_capability" "existing" {
  slug = "incident-responder"
}

output "capability_instructions" {
  value = data.roadie_capability.existing.instructions
}
