# Import an existing Opsgenie notifier using its UUID
terraform import oodle_notifier.platform_opsgenie 123e4567-e89b-12d3-a456-426614174001

# Import an existing Slack notifier for critical alerts using its UUID
terraform import oodle_notifier.critical_slack 123e4567-e89b-12d3-a456-426614174002

# Import an existing Slack notifier for general alerts using its UUID
terraform import oodle_notifier.general_slack 123e4567-e89b-12d3-a456-426614174003
