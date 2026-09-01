resource "roadie_secret" "pagerduty_token" {
  ref         = "PAGERDUTY_TOKEN"
  name        = "PagerDuty Token"
  description = "API token for PagerDuty integration"
  value       = var.pagerduty_token
}
