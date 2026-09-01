resource "roadie_relationship_rule" "repo_to_pagerduty" {
  name                    = "GitHub Repo to PagerDuty Service"
  description             = "Links GitHub repositories to their PagerDuty services by name."
  source_datasource_id    = roadie_data_source.github_repos.id
  target_datasource_id    = roadie_data_source.pagerduty_services.id
  source_field_expression = "$.name"
  target_field_expression = "$.service.name"
  relationship_type       = "monitoredBy"
  strategy                = "field-matching"
  match_strategy          = "exact"
  state                   = "active"
}
