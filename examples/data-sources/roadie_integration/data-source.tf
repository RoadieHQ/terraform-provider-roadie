data "roadie_integration" "github" {
  slug = "github"
}

output "github_host" {
  value = data.roadie_integration.github.host
}
