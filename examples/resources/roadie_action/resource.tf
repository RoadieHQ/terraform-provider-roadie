resource "roadie_action" "list_repos" {
  name        = "List Repositories"
  slug        = "list-repositories"
  description = "Lists repositories from GitHub."
  enabled     = true

  parameters = [
    {
      name        = "org"
      type        = "string"
      description = "The GitHub organization to list repositories for."
      required    = true
    },
  ]

  steps = [
    {
      id             = "fetch_repos"
      integration_id = roadie_integration.github.id
      request = jsonencode({
        method  = "GET"
        path    = "/orgs/{{org}}/repos"
        headers = [{ key = "Accept", value = "application/json" }]
        body    = ""
      })
    },
  ]
}
