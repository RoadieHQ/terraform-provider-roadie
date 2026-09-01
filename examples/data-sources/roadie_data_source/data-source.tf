data "roadie_data_source" "existing" {
  id = "data-source-uuid-here"
}

output "data_source_enabled" {
  value = data.roadie_data_source.existing.enabled
}
