package constants

const (
	OpenState   = "open"
	ClosedState = "closed"

	UpdatedSortOption   = "updated"
	DescDirectionOption = "desc"

	GithubTokenEnvVar = "GITHUB_TOKEN"
	GitlabTokenEnvVar = "GITLAB_TOKEN"

	// https://docs.gitlab.com/ee/user/gitlab_com/index.html#gitlabcom-specific-rate-limits
	GitlabRateLimitHeader          = "RateLimit-Limit"
	GitlabRateLimitRemainingHeader = "RateLimit-Remaining"
	GitlabRateLimitResetHeader     = "RateLimit-Reset"
)
