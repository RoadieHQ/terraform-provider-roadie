resource "roadie_context_group" "services_and_teams" {
  name        = "Services and Teams"
  slug        = "services-and-teams"
  description = "Maps services to their owning teams."

  datasources = [
    {
      datasource_id = roadie_data_source.github_repos.id
    },
    {
      datasource_id = roadie_data_source.pagerduty_services.id
      filter        = "[{\"field\":\"status\",\"operator\":\"equals\",\"value\":\"active\"}]"
    },
  ]

  merge_relationship_types   = ["ownedBy"]
  include_external_relations = true
}
