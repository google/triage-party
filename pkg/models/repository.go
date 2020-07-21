package models

// Repository represents a GitHub repository.
type Repository struct {
	ID          *int64  `json:"id,omitempty"`
	NodeID      *string `json:"node_id,omitempty"`
	Owner       *User   `json:"owner,omitempty"`
	Name        *string `json:"name,omitempty"`
	FullName    *string `json:"full_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Homepage    *string `json:"homepage,omitempty"`
	//CodeOfConduct       *CodeOfConduct   `json:"code_of_conduct,omitempty"`
	DefaultBranch      *string     `json:"default_branch,omitempty"`
	MasterBranch       *string     `json:"master_branch,omitempty"`
	CreatedAt          *Timestamp  `json:"created_at,omitempty"`
	PushedAt           *Timestamp  `json:"pushed_at,omitempty"`
	UpdatedAt          *Timestamp  `json:"updated_at,omitempty"`
	HTMLURL            *string     `json:"html_url,omitempty"`
	CloneURL           *string     `json:"clone_url,omitempty"`
	GitURL             *string     `json:"git_url,omitempty"`
	MirrorURL          *string     `json:"mirror_url,omitempty"`
	SSHURL             *string     `json:"ssh_url,omitempty"`
	SVNURL             *string     `json:"svn_url,omitempty"`
	Language           *string     `json:"language,omitempty"`
	Fork               *bool       `json:"fork,omitempty"`
	ForksCount         *int        `json:"forks_count,omitempty"`
	NetworkCount       *int        `json:"network_count,omitempty"`
	OpenIssuesCount    *int        `json:"open_issues_count,omitempty"`
	StargazersCount    *int        `json:"stargazers_count,omitempty"`
	SubscribersCount   *int        `json:"subscribers_count,omitempty"`
	WatchersCount      *int        `json:"watchers_count,omitempty"`
	Size               *int        `json:"size,omitempty"`
	AutoInit           *bool       `json:"auto_init,omitempty"`
	Parent             *Repository `json:"parent,omitempty"`
	Source             *Repository `json:"source,omitempty"`
	TemplateRepository *Repository `json:"template_repository,omitempty"`
	//Organization        *Organization    `json:"organization,omitempty"`
	Permissions         *map[string]bool `json:"permissions,omitempty"`
	AllowRebaseMerge    *bool            `json:"allow_rebase_merge,omitempty"`
	AllowSquashMerge    *bool            `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool            `json:"allow_merge_commit,omitempty"`
	DeleteBranchOnMerge *bool            `json:"delete_branch_on_merge,omitempty"`
	Topics              []string         `json:"topics,omitempty"`
	Archived            *bool            `json:"archived,omitempty"`
	Disabled            *bool            `json:"disabled,omitempty"`

	// Only provided when using RepositoriesService.Get while in preview
	//License *License `json:"license,omitempty"`

	// Additional mutable fields when creating and editing a repository
	Private           *bool   `json:"private,omitempty"`
	HasIssues         *bool   `json:"has_issues,omitempty"`
	HasWiki           *bool   `json:"has_wiki,omitempty"`
	HasPages          *bool   `json:"has_pages,omitempty"`
	HasProjects       *bool   `json:"has_projects,omitempty"`
	HasDownloads      *bool   `json:"has_downloads,omitempty"`
	IsTemplate        *bool   `json:"is_template,omitempty"`
	LicenseTemplate   *string `json:"license_template,omitempty"`
	GitignoreTemplate *string `json:"gitignore_template,omitempty"`

	// Creating an organization repository. Required for non-owners.
	TeamID *int64 `json:"team_id,omitempty"`

	// API URLs
	URL              *string `json:"url,omitempty"`
	ArchiveURL       *string `json:"archive_url,omitempty"`
	AssigneesURL     *string `json:"assignees_url,omitempty"`
	BlobsURL         *string `json:"blobs_url,omitempty"`
	BranchesURL      *string `json:"branches_url,omitempty"`
	CollaboratorsURL *string `json:"collaborators_url,omitempty"`
	CommentsURL      *string `json:"comments_url,omitempty"`
	CommitsURL       *string `json:"commits_url,omitempty"`
	CompareURL       *string `json:"compare_url,omitempty"`
	ContentsURL      *string `json:"contents_url,omitempty"`
	ContributorsURL  *string `json:"contributors_url,omitempty"`
	DeploymentsURL   *string `json:"deployments_url,omitempty"`
	DownloadsURL     *string `json:"downloads_url,omitempty"`
	EventsURL        *string `json:"events_url,omitempty"`
	ForksURL         *string `json:"forks_url,omitempty"`
	GitCommitsURL    *string `json:"git_commits_url,omitempty"`
	GitRefsURL       *string `json:"git_refs_url,omitempty"`
	GitTagsURL       *string `json:"git_tags_url,omitempty"`
	HooksURL         *string `json:"hooks_url,omitempty"`
	IssueCommentURL  *string `json:"issue_comment_url,omitempty"`
	IssueEventsURL   *string `json:"issue_events_url,omitempty"`
	IssuesURL        *string `json:"issues_url,omitempty"`
	KeysURL          *string `json:"keys_url,omitempty"`
	LabelsURL        *string `json:"labels_url,omitempty"`
	LanguagesURL     *string `json:"languages_url,omitempty"`
	MergesURL        *string `json:"merges_url,omitempty"`
	MilestonesURL    *string `json:"milestones_url,omitempty"`
	NotificationsURL *string `json:"notifications_url,omitempty"`
	PullsURL         *string `json:"pulls_url,omitempty"`
	ReleasesURL      *string `json:"releases_url,omitempty"`
	StargazersURL    *string `json:"stargazers_url,omitempty"`
	StatusesURL      *string `json:"statuses_url,omitempty"`
	SubscribersURL   *string `json:"subscribers_url,omitempty"`
	SubscriptionURL  *string `json:"subscription_url,omitempty"`
	TagsURL          *string `json:"tags_url,omitempty"`
	TreesURL         *string `json:"trees_url,omitempty"`
	TeamsURL         *string `json:"teams_url,omitempty"`

	// TextMatches is only populated from search results that request text matches
	// See: search.go and https://developer.github.com/v3/search/#text-match-metadata
	//TextMatches []*TextMatch `json:"text_matches,omitempty"`

	// Visibility is only used for Create and Edit endpoints. The visibility field
	// overrides the field parameter when both are used.
	// Can be one of public, private or internal.
	Visibility *string `json:"visibility,omitempty"`
}

// GetAllowMergeCommit returns the AllowMergeCommit field if it's non-nil, zero value otherwise.
func (r *Repository) GetAllowMergeCommit() bool {
	if r == nil || r.AllowMergeCommit == nil {
		return false
	}
	return *r.AllowMergeCommit
}

// GetAllowRebaseMerge returns the AllowRebaseMerge field if it's non-nil, zero value otherwise.
func (r *Repository) GetAllowRebaseMerge() bool {
	if r == nil || r.AllowRebaseMerge == nil {
		return false
	}
	return *r.AllowRebaseMerge
}

// GetAllowSquashMerge returns the AllowSquashMerge field if it's non-nil, zero value otherwise.
func (r *Repository) GetAllowSquashMerge() bool {
	if r == nil || r.AllowSquashMerge == nil {
		return false
	}
	return *r.AllowSquashMerge
}

// GetArchived returns the Archived field if it's non-nil, zero value otherwise.
func (r *Repository) GetArchived() bool {
	if r == nil || r.Archived == nil {
		return false
	}
	return *r.Archived
}

// GetArchiveURL returns the ArchiveURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetArchiveURL() string {
	if r == nil || r.ArchiveURL == nil {
		return ""
	}
	return *r.ArchiveURL
}

// GetAssigneesURL returns the AssigneesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetAssigneesURL() string {
	if r == nil || r.AssigneesURL == nil {
		return ""
	}
	return *r.AssigneesURL
}

// GetAutoInit returns the AutoInit field if it's non-nil, zero value otherwise.
func (r *Repository) GetAutoInit() bool {
	if r == nil || r.AutoInit == nil {
		return false
	}
	return *r.AutoInit
}

// GetBlobsURL returns the BlobsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetBlobsURL() string {
	if r == nil || r.BlobsURL == nil {
		return ""
	}
	return *r.BlobsURL
}

// GetBranchesURL returns the BranchesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetBranchesURL() string {
	if r == nil || r.BranchesURL == nil {
		return ""
	}
	return *r.BranchesURL
}

// GetCloneURL returns the CloneURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetCloneURL() string {
	if r == nil || r.CloneURL == nil {
		return ""
	}
	return *r.CloneURL
}

// GetCodeOfConduct returns the CodeOfConduct field.
//func (r *Repository) GetCodeOfConduct() *CodeOfConduct {
//	if r == nil {
//		return nil
//	}
//	return r.CodeOfConduct
//}

// GetCollaboratorsURL returns the CollaboratorsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetCollaboratorsURL() string {
	if r == nil || r.CollaboratorsURL == nil {
		return ""
	}
	return *r.CollaboratorsURL
}

// GetCommentsURL returns the CommentsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetCommentsURL() string {
	if r == nil || r.CommentsURL == nil {
		return ""
	}
	return *r.CommentsURL
}

// GetCommitsURL returns the CommitsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetCommitsURL() string {
	if r == nil || r.CommitsURL == nil {
		return ""
	}
	return *r.CommitsURL
}

// GetCompareURL returns the CompareURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetCompareURL() string {
	if r == nil || r.CompareURL == nil {
		return ""
	}
	return *r.CompareURL
}

// GetContentsURL returns the ContentsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetContentsURL() string {
	if r == nil || r.ContentsURL == nil {
		return ""
	}
	return *r.ContentsURL
}

// GetContributorsURL returns the ContributorsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetContributorsURL() string {
	if r == nil || r.ContributorsURL == nil {
		return ""
	}
	return *r.ContributorsURL
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (r *Repository) GetCreatedAt() Timestamp {
	if r == nil || r.CreatedAt == nil {
		return Timestamp{}
	}
	return *r.CreatedAt
}

// GetDefaultBranch returns the DefaultBranch field if it's non-nil, zero value otherwise.
func (r *Repository) GetDefaultBranch() string {
	if r == nil || r.DefaultBranch == nil {
		return ""
	}
	return *r.DefaultBranch
}

// GetDeleteBranchOnMerge returns the DeleteBranchOnMerge field if it's non-nil, zero value otherwise.
func (r *Repository) GetDeleteBranchOnMerge() bool {
	if r == nil || r.DeleteBranchOnMerge == nil {
		return false
	}
	return *r.DeleteBranchOnMerge
}

// GetDeploymentsURL returns the DeploymentsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetDeploymentsURL() string {
	if r == nil || r.DeploymentsURL == nil {
		return ""
	}
	return *r.DeploymentsURL
}

// GetDescription returns the Description field if it's non-nil, zero value otherwise.
func (r *Repository) GetDescription() string {
	if r == nil || r.Description == nil {
		return ""
	}
	return *r.Description
}

// GetDisabled returns the Disabled field if it's non-nil, zero value otherwise.
func (r *Repository) GetDisabled() bool {
	if r == nil || r.Disabled == nil {
		return false
	}
	return *r.Disabled
}

// GetDownloadsURL returns the DownloadsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetDownloadsURL() string {
	if r == nil || r.DownloadsURL == nil {
		return ""
	}
	return *r.DownloadsURL
}

// GetEventsURL returns the EventsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetEventsURL() string {
	if r == nil || r.EventsURL == nil {
		return ""
	}
	return *r.EventsURL
}

// GetFork returns the Fork field if it's non-nil, zero value otherwise.
func (r *Repository) GetFork() bool {
	if r == nil || r.Fork == nil {
		return false
	}
	return *r.Fork
}

// GetForksCount returns the ForksCount field if it's non-nil, zero value otherwise.
func (r *Repository) GetForksCount() int {
	if r == nil || r.ForksCount == nil {
		return 0
	}
	return *r.ForksCount
}

// GetForksURL returns the ForksURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetForksURL() string {
	if r == nil || r.ForksURL == nil {
		return ""
	}
	return *r.ForksURL
}

// GetFullName returns the FullName field if it's non-nil, zero value otherwise.
func (r *Repository) GetFullName() string {
	if r == nil || r.FullName == nil {
		return ""
	}
	return *r.FullName
}

// GetGitCommitsURL returns the GitCommitsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetGitCommitsURL() string {
	if r == nil || r.GitCommitsURL == nil {
		return ""
	}
	return *r.GitCommitsURL
}

// GetGitignoreTemplate returns the GitignoreTemplate field if it's non-nil, zero value otherwise.
func (r *Repository) GetGitignoreTemplate() string {
	if r == nil || r.GitignoreTemplate == nil {
		return ""
	}
	return *r.GitignoreTemplate
}

// GetGitRefsURL returns the GitRefsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetGitRefsURL() string {
	if r == nil || r.GitRefsURL == nil {
		return ""
	}
	return *r.GitRefsURL
}

// GetGitTagsURL returns the GitTagsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetGitTagsURL() string {
	if r == nil || r.GitTagsURL == nil {
		return ""
	}
	return *r.GitTagsURL
}

// GetGitURL returns the GitURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetGitURL() string {
	if r == nil || r.GitURL == nil {
		return ""
	}
	return *r.GitURL
}

// GetHasDownloads returns the HasDownloads field if it's non-nil, zero value otherwise.
func (r *Repository) GetHasDownloads() bool {
	if r == nil || r.HasDownloads == nil {
		return false
	}
	return *r.HasDownloads
}

// GetHasIssues returns the HasIssues field if it's non-nil, zero value otherwise.
func (r *Repository) GetHasIssues() bool {
	if r == nil || r.HasIssues == nil {
		return false
	}
	return *r.HasIssues
}

// GetHasPages returns the HasPages field if it's non-nil, zero value otherwise.
func (r *Repository) GetHasPages() bool {
	if r == nil || r.HasPages == nil {
		return false
	}
	return *r.HasPages
}

// GetHasProjects returns the HasProjects field if it's non-nil, zero value otherwise.
func (r *Repository) GetHasProjects() bool {
	if r == nil || r.HasProjects == nil {
		return false
	}
	return *r.HasProjects
}

// GetHasWiki returns the HasWiki field if it's non-nil, zero value otherwise.
func (r *Repository) GetHasWiki() bool {
	if r == nil || r.HasWiki == nil {
		return false
	}
	return *r.HasWiki
}

// GetHomepage returns the Homepage field if it's non-nil, zero value otherwise.
func (r *Repository) GetHomepage() string {
	if r == nil || r.Homepage == nil {
		return ""
	}
	return *r.Homepage
}

// GetHooksURL returns the HooksURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetHooksURL() string {
	if r == nil || r.HooksURL == nil {
		return ""
	}
	return *r.HooksURL
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetHTMLURL() string {
	if r == nil || r.HTMLURL == nil {
		return ""
	}
	return *r.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (r *Repository) GetID() int64 {
	if r == nil || r.ID == nil {
		return 0
	}
	return *r.ID
}

// GetIssueCommentURL returns the IssueCommentURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetIssueCommentURL() string {
	if r == nil || r.IssueCommentURL == nil {
		return ""
	}
	return *r.IssueCommentURL
}

// GetIssueEventsURL returns the IssueEventsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetIssueEventsURL() string {
	if r == nil || r.IssueEventsURL == nil {
		return ""
	}
	return *r.IssueEventsURL
}

// GetIssuesURL returns the IssuesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetIssuesURL() string {
	if r == nil || r.IssuesURL == nil {
		return ""
	}
	return *r.IssuesURL
}

// GetIsTemplate returns the IsTemplate field if it's non-nil, zero value otherwise.
func (r *Repository) GetIsTemplate() bool {
	if r == nil || r.IsTemplate == nil {
		return false
	}
	return *r.IsTemplate
}

// GetKeysURL returns the KeysURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetKeysURL() string {
	if r == nil || r.KeysURL == nil {
		return ""
	}
	return *r.KeysURL
}

// GetLabelsURL returns the LabelsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetLabelsURL() string {
	if r == nil || r.LabelsURL == nil {
		return ""
	}
	return *r.LabelsURL
}

// GetLanguage returns the Language field if it's non-nil, zero value otherwise.
func (r *Repository) GetLanguage() string {
	if r == nil || r.Language == nil {
		return ""
	}
	return *r.Language
}

// GetLanguagesURL returns the LanguagesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetLanguagesURL() string {
	if r == nil || r.LanguagesURL == nil {
		return ""
	}
	return *r.LanguagesURL
}

// GetLicense returns the License field.
//func (r *Repository) GetLicense() *License {
//	if r == nil {
//		return nil
//	}
//	return r.License
//}

// GetLicenseTemplate returns the LicenseTemplate field if it's non-nil, zero value otherwise.
func (r *Repository) GetLicenseTemplate() string {
	if r == nil || r.LicenseTemplate == nil {
		return ""
	}
	return *r.LicenseTemplate
}

// GetMasterBranch returns the MasterBranch field if it's non-nil, zero value otherwise.
func (r *Repository) GetMasterBranch() string {
	if r == nil || r.MasterBranch == nil {
		return ""
	}
	return *r.MasterBranch
}

// GetMergesURL returns the MergesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetMergesURL() string {
	if r == nil || r.MergesURL == nil {
		return ""
	}
	return *r.MergesURL
}

// GetMilestonesURL returns the MilestonesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetMilestonesURL() string {
	if r == nil || r.MilestonesURL == nil {
		return ""
	}
	return *r.MilestonesURL
}

// GetMirrorURL returns the MirrorURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetMirrorURL() string {
	if r == nil || r.MirrorURL == nil {
		return ""
	}
	return *r.MirrorURL
}

// GetName returns the Name field if it's non-nil, zero value otherwise.
func (r *Repository) GetName() string {
	if r == nil || r.Name == nil {
		return ""
	}
	return *r.Name
}

// GetNetworkCount returns the NetworkCount field if it's non-nil, zero value otherwise.
func (r *Repository) GetNetworkCount() int {
	if r == nil || r.NetworkCount == nil {
		return 0
	}
	return *r.NetworkCount
}

// GetNodeID returns the NodeID field if it's non-nil, zero value otherwise.
func (r *Repository) GetNodeID() string {
	if r == nil || r.NodeID == nil {
		return ""
	}
	return *r.NodeID
}

// GetNotificationsURL returns the NotificationsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetNotificationsURL() string {
	if r == nil || r.NotificationsURL == nil {
		return ""
	}
	return *r.NotificationsURL
}

// GetOpenIssuesCount returns the OpenIssuesCount field if it's non-nil, zero value otherwise.
func (r *Repository) GetOpenIssuesCount() int {
	if r == nil || r.OpenIssuesCount == nil {
		return 0
	}
	return *r.OpenIssuesCount
}

// GetOrganization returns the Organization field.
//func (r *Repository) GetOrganization() *Organization {
//	if r == nil {
//		return nil
//	}
//	return r.Organization
//}

// GetOwner returns the Owner field.
func (r *Repository) GetOwner() *User {
	if r == nil {
		return nil
	}
	return r.Owner
}

// GetParent returns the Parent field.
func (r *Repository) GetParent() *Repository {
	if r == nil {
		return nil
	}
	return r.Parent
}

// GetPermissions returns the Permissions field if it's non-nil, zero value otherwise.
func (r *Repository) GetPermissions() map[string]bool {
	if r == nil || r.Permissions == nil {
		return map[string]bool{}
	}
	return *r.Permissions
}

// GetPrivate returns the Private field if it's non-nil, zero value otherwise.
func (r *Repository) GetPrivate() bool {
	if r == nil || r.Private == nil {
		return false
	}
	return *r.Private
}

// GetPullsURL returns the PullsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetPullsURL() string {
	if r == nil || r.PullsURL == nil {
		return ""
	}
	return *r.PullsURL
}

// GetPushedAt returns the PushedAt field if it's non-nil, zero value otherwise.
func (r *Repository) GetPushedAt() Timestamp {
	if r == nil || r.PushedAt == nil {
		return Timestamp{}
	}
	return *r.PushedAt
}

// GetReleasesURL returns the ReleasesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetReleasesURL() string {
	if r == nil || r.ReleasesURL == nil {
		return ""
	}
	return *r.ReleasesURL
}

// GetSize returns the Size field if it's non-nil, zero value otherwise.
func (r *Repository) GetSize() int {
	if r == nil || r.Size == nil {
		return 0
	}
	return *r.Size
}

// GetSource returns the Source field.
func (r *Repository) GetSource() *Repository {
	if r == nil {
		return nil
	}
	return r.Source
}

// GetSSHURL returns the SSHURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetSSHURL() string {
	if r == nil || r.SSHURL == nil {
		return ""
	}
	return *r.SSHURL
}

// GetStargazersCount returns the StargazersCount field if it's non-nil, zero value otherwise.
func (r *Repository) GetStargazersCount() int {
	if r == nil || r.StargazersCount == nil {
		return 0
	}
	return *r.StargazersCount
}

// GetStargazersURL returns the StargazersURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetStargazersURL() string {
	if r == nil || r.StargazersURL == nil {
		return ""
	}
	return *r.StargazersURL
}

// GetStatusesURL returns the StatusesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetStatusesURL() string {
	if r == nil || r.StatusesURL == nil {
		return ""
	}
	return *r.StatusesURL
}

// GetSubscribersCount returns the SubscribersCount field if it's non-nil, zero value otherwise.
func (r *Repository) GetSubscribersCount() int {
	if r == nil || r.SubscribersCount == nil {
		return 0
	}
	return *r.SubscribersCount
}

// GetSubscribersURL returns the SubscribersURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetSubscribersURL() string {
	if r == nil || r.SubscribersURL == nil {
		return ""
	}
	return *r.SubscribersURL
}

// GetSubscriptionURL returns the SubscriptionURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetSubscriptionURL() string {
	if r == nil || r.SubscriptionURL == nil {
		return ""
	}
	return *r.SubscriptionURL
}

// GetSVNURL returns the SVNURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetSVNURL() string {
	if r == nil || r.SVNURL == nil {
		return ""
	}
	return *r.SVNURL
}

// GetTagsURL returns the TagsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetTagsURL() string {
	if r == nil || r.TagsURL == nil {
		return ""
	}
	return *r.TagsURL
}

// GetTeamID returns the TeamID field if it's non-nil, zero value otherwise.
func (r *Repository) GetTeamID() int64 {
	if r == nil || r.TeamID == nil {
		return 0
	}
	return *r.TeamID
}

// GetTeamsURL returns the TeamsURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetTeamsURL() string {
	if r == nil || r.TeamsURL == nil {
		return ""
	}
	return *r.TeamsURL
}

// GetTemplateRepository returns the TemplateRepository field.
func (r *Repository) GetTemplateRepository() *Repository {
	if r == nil {
		return nil
	}
	return r.TemplateRepository
}

// GetTreesURL returns the TreesURL field if it's non-nil, zero value otherwise.
func (r *Repository) GetTreesURL() string {
	if r == nil || r.TreesURL == nil {
		return ""
	}
	return *r.TreesURL
}

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (r *Repository) GetUpdatedAt() Timestamp {
	if r == nil || r.UpdatedAt == nil {
		return Timestamp{}
	}
	return *r.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (r *Repository) GetURL() string {
	if r == nil || r.URL == nil {
		return ""
	}
	return *r.URL
}

// GetVisibility returns the Visibility field if it's non-nil, zero value otherwise.
func (r *Repository) GetVisibility() string {
	if r == nil || r.Visibility == nil {
		return ""
	}
	return *r.Visibility
}

// GetWatchersCount returns the WatchersCount field if it's non-nil, zero value otherwise.
func (r *Repository) GetWatchersCount() int {
	if r == nil || r.WatchersCount == nil {
		return 0
	}
	return *r.WatchersCount
}
