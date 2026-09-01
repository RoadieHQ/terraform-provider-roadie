resource "roadie_secret" "github_token" {
  ref         = "GITHUB_TOKEN"
  name        = "GitHub Token"
  description = "Personal access token for GitHub API"
  value       = var.github_token
}

resource "roadie_integration" "github" {
  name         = "GitHub"
  slug         = "github"
  type         = "scm"
  host         = "https://api.github.com"
  auth_type    = "bearer-token"
  backend_type = "http"
  enabled      = true

  auth_config = jsonencode({
    token = "$${GITHUB_TOKEN}"
  })

  depends_on = [roadie_secret.github_token]
}
