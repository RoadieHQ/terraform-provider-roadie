resource "roadie_capability" "example" {
  name         = "Incident Responder"
  description  = "Handles incident triage and response."
  instructions = <<-EOT
    You are an incident responder. When triggered, assess the severity
    of the incident and recommend appropriate actions.
  EOT
}
