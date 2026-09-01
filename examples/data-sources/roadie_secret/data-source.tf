data "roadie_secret" "github_token" {
  ref = "GITHUB_TOKEN"
}

output "github_token_status" {
  value = data.roadie_secret.github_token.status
}
