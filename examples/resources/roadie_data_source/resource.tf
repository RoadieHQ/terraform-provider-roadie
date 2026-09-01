resource "roadie_data_source" "github_repos" {
  name          = "GitHub Repositories"
  slug          = "github-repositories"
  description   = "Ingests repositories from GitHub."
  workflow_type = "data-ingestion"
  enabled       = true

  nodes = jsonencode([
    {
      id       = "trigger-node"
      type     = "trigger-schedule"
      position = { x = 0, y = 0 }
      data = {
        label  = "Schedule"
        config = { frequencyValue = 1, frequencyUnit = "days" }
      }
    },
    {
      id       = "integration-node"
      type     = "source-integration"
      position = { x = 300, y = 0 }
      data = {
        label  = "GitHub"
        config = {
          integrationId = roadie_integration.github.id
          method        = "GET"
          path          = "/orgs/my-org/repos"
        }
      }
    },
    {
      id       = "sink-node"
      type     = "sink-datastore"
      position = { x = 600, y = 0 }
      data = {
        label  = "Datastore"
        config = { id_selector = "full_name" }
      }
    },
  ])

  edges = jsonencode([
    {
      id     = "e1"
      source = "trigger-node"
      target = "integration-node"
    },
    {
      id     = "e2"
      source = "integration-node"
      target = "sink-node"
    },
  ])
}
