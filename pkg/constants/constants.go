package constants

const (
	OpenState   = "open"
	OpenedState = "opened"
	ClosedState = "closed"

	UpdatedSortOption   = "updated"
	UpdatedAtSortOption = "updated_at"
	CreatedAtSortOption = "created_at"
	DescDirectionOption = "desc"

	GithubTokenEnvVar = "GITHUB_TOKEN"
	GitlabTokenEnvVar = "GITLAB_TOKEN"

	GithubProviderName = "github"
	GitlabProviderName = "gitlab"

	// https://docs.gitlab.com/ee/user/gitlab_com/index.html#gitlabcom-specific-rate-limits
	GitlabRateLimitHeader          = "RateLimit-Limit"
	GitlabRateLimitRemainingHeader = "RateLimit-Remaining"
	GitlabRateLimitResetHeader     = "RateLimit-Reset"

	GithubProviderHost = "github.com"
	GitlabProviderHost = "gitlab.com"
)
