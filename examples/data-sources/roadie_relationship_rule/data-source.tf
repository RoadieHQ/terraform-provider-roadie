data "roadie_relationship_rule" "ownership" {
  id = "rule-uuid-here"
}

output "rule_state" {
  value = data.roadie_relationship_rule.ownership.state
}
