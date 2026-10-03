# Import the advanced settings of a space by its space ID.
terraform import elasticstack_kibana_advanced_settings.team_a team-a

# Import the global advanced settings.
terraform import elasticstack_kibana_advanced_settings.global global

# Only the scope is imported. Declare the settings to manage in the
# configuration before running terraform plan / apply.
