# A dataset carries at most one schedule, so it is imported by the
# dataset's name rather than by the schedule's own uuid.
terraform import oodle_genai_dataset_schedule.support_eval_weekdays support-eval

# The first plan after an import may show a change in experiment_config
# (whitespace only, because the server stores compact JSON) and in mode
# when the config states the default "calendar". One apply settles both.
