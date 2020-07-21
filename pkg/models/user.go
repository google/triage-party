package models

// abstraction model for github.User struct
// User represents a GitHub user.
type User struct {
	Login                   *string    `json:"login,omitempty"`
	ID                      *int64     `json:"id,omitempty"`
	NodeID                  *string    `json:"node_id,omitempty"`
	AvatarURL               *string    `json:"avatar_url,omitempty"`
	HTMLURL                 *string    `json:"html_url,omitempty"`
	GravatarID              *string    `json:"gravatar_id,omitempty"`
	Name                    *string    `json:"name,omitempty"`
	Company                 *string    `json:"company,omitempty"`
	Blog                    *string    `json:"blog,omitempty"`
	Location                *string    `json:"location,omitempty"`
	Email                   *string    `json:"email,omitempty"`
	Hireable                *bool      `json:"hireable,omitempty"`
	Bio                     *string    `json:"bio,omitempty"`
	PublicRepos             *int       `json:"public_repos,omitempty"`
	PublicGists             *int       `json:"public_gists,omitempty"`
	Followers               *int       `json:"followers,omitempty"`
	Following               *int       `json:"following,omitempty"`
	CreatedAt               *Timestamp `json:"created_at,omitempty"`
	UpdatedAt               *Timestamp `json:"updated_at,omitempty"`
	SuspendedAt             *Timestamp `json:"suspended_at,omitempty"`
	Type                    *string    `json:"type,omitempty"`
	SiteAdmin               *bool      `json:"site_admin,omitempty"`
	TotalPrivateRepos       *int       `json:"total_private_repos,omitempty"`
	OwnedPrivateRepos       *int       `json:"owned_private_repos,omitempty"`
	PrivateGists            *int       `json:"private_gists,omitempty"`
	DiskUsage               *int       `json:"disk_usage,omitempty"`
	Collaborators           *int       `json:"collaborators,omitempty"`
	TwoFactorAuthentication *bool      `json:"two_factor_authentication,omitempty"`

	// TODO do we need it?
	//Plan                    *Plan      `json:"plan,omitempty"`

	LdapDn *string `json:"ldap_dn,omitempty"`

	// API URLs
	URL               *string `json:"url,omitempty"`
	EventsURL         *string `json:"events_url,omitempty"`
	FollowingURL      *string `json:"following_url,omitempty"`
	FollowersURL      *string `json:"followers_url,omitempty"`
	GistsURL          *string `json:"gists_url,omitempty"`
	OrganizationsURL  *string `json:"organizations_url,omitempty"`
	ReceivedEventsURL *string `json:"received_events_url,omitempty"`
	ReposURL          *string `json:"repos_url,omitempty"`
	StarredURL        *string `json:"starred_url,omitempty"`
	SubscriptionsURL  *string `json:"subscriptions_url,omitempty"`

	// TODO do we need it?
	// TextMatches is only populated from search results that request text matches
	// See: search.go and https://developer.github.com/v3/search/#text-match-metadata
	//TextMatches []*TextMatch `json:"text_matches,omitempty"`

	// Permissions identifies the permissions that a user has on a given
	// repository. This is only populated when calling Repositories.ListCollaborators.
	Permissions *map[string]bool `json:"permissions,omitempty"`
}

// GetAvatarURL returns the AvatarURL field if it's non-nil, zero value otherwise.
func (u *User) GetAvatarURL() string {
	if u == nil || u.AvatarURL == nil {
		return ""
	}
	return *u.AvatarURL
}

// GetBio returns the Bio field if it's non-nil, zero value otherwise.
func (u *User) GetBio() string {
	if u == nil || u.Bio == nil {
		return ""
	}
	return *u.Bio
}

// GetBlog returns the Blog field if it's non-nil, zero value otherwise.
func (u *User) GetBlog() string {
	if u == nil || u.Blog == nil {
		return ""
	}
	return *u.Blog
}

// GetCollaborators returns the Collaborators field if it's non-nil, zero value otherwise.
func (u *User) GetCollaborators() int {
	if u == nil || u.Collaborators == nil {
		return 0
	}
	return *u.Collaborators
}

// GetCompany returns the Company field if it's non-nil, zero value otherwise.
func (u *User) GetCompany() string {
	if u == nil || u.Company == nil {
		return ""
	}
	return *u.Company
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (u *User) GetCreatedAt() Timestamp {
	if u == nil || u.CreatedAt == nil {
		return Timestamp{}
	}
	return *u.CreatedAt
}

// GetDiskUsage returns the DiskUsage field if it's non-nil, zero value otherwise.
func (u *User) GetDiskUsage() int {
	if u == nil || u.DiskUsage == nil {
		return 0
	}
	return *u.DiskUsage
}

// GetEmail returns the Email field if it's non-nil, zero value otherwise.
func (u *User) GetEmail() string {
	if u == nil || u.Email == nil {
		return ""
	}
	return *u.Email
}

// GetEventsURL returns the EventsURL field if it's non-nil, zero value otherwise.
func (u *User) GetEventsURL() string {
	if u == nil || u.EventsURL == nil {
		return ""
	}
	return *u.EventsURL
}

// GetFollowers returns the Followers field if it's non-nil, zero value otherwise.
func (u *User) GetFollowers() int {
	if u == nil || u.Followers == nil {
		return 0
	}
	return *u.Followers
}

// GetFollowersURL returns the FollowersURL field if it's non-nil, zero value otherwise.
func (u *User) GetFollowersURL() string {
	if u == nil || u.FollowersURL == nil {
		return ""
	}
	return *u.FollowersURL
}

// GetFollowing returns the Following field if it's non-nil, zero value otherwise.
func (u *User) GetFollowing() int {
	if u == nil || u.Following == nil {
		return 0
	}
	return *u.Following
}

// GetFollowingURL returns the FollowingURL field if it's non-nil, zero value otherwise.
func (u *User) GetFollowingURL() string {
	if u == nil || u.FollowingURL == nil {
		return ""
	}
	return *u.FollowingURL
}

// GetGistsURL returns the GistsURL field if it's non-nil, zero value otherwise.
func (u *User) GetGistsURL() string {
	if u == nil || u.GistsURL == nil {
		return ""
	}
	return *u.GistsURL
}

// GetGravatarID returns the GravatarID field if it's non-nil, zero value otherwise.
func (u *User) GetGravatarID() string {
	if u == nil || u.GravatarID == nil {
		return ""
	}
	return *u.GravatarID
}

// GetHireable returns the Hireable field if it's non-nil, zero value otherwise.
func (u *User) GetHireable() bool {
	if u == nil || u.Hireable == nil {
		return false
	}
	return *u.Hireable
}

// GetHTMLURL returns the HTMLURL field if it's non-nil, zero value otherwise.
func (u *User) GetHTMLURL() string {
	if u == nil || u.HTMLURL == nil {
		return ""
	}
	return *u.HTMLURL
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (u *User) GetID() int64 {
	if u == nil || u.ID == nil {
		return 0
	}
	return *u.ID
}

// GetLdapDn returns the LdapDn field if it's non-nil, zero value otherwise.
func (u *User) GetLdapDn() string {
	if u == nil || u.LdapDn == nil {
		return ""
	}
	return *u.LdapDn
}

// GetLocation returns the Location field if it's non-nil, zero value otherwise.
func (u *User) GetLocation() string {
	if u == nil || u.Location == nil {
		return ""
	}
	return *u.Location
}

// GetLogin returns the Login field if it's non-nil, zero value otherwise.
func (u *User) GetLogin() string {
	if u == nil || u.Login == nil {
		return ""
	}
	return *u.Login
}

// GetName returns the Name field if it's non-nil, zero value otherwise.
func (u *User) GetName() string {
	if u == nil || u.Name == nil {
		return ""
	}
	return *u.Name
}

// GetNodeID returns the NodeID field if it's non-nil, zero value otherwise.
func (u *User) GetNodeID() string {
	if u == nil || u.NodeID == nil {
		return ""
	}
	return *u.NodeID
}

// GetOrganizationsURL returns the OrganizationsURL field if it's non-nil, zero value otherwise.
func (u *User) GetOrganizationsURL() string {
	if u == nil || u.OrganizationsURL == nil {
		return ""
	}
	return *u.OrganizationsURL
}

// GetOwnedPrivateRepos returns the OwnedPrivateRepos field if it's non-nil, zero value otherwise.
func (u *User) GetOwnedPrivateRepos() int {
	if u == nil || u.OwnedPrivateRepos == nil {
		return 0
	}
	return *u.OwnedPrivateRepos
}

// GetPermissions returns the Permissions field if it's non-nil, zero value otherwise.
func (u *User) GetPermissions() map[string]bool {
	if u == nil || u.Permissions == nil {
		return map[string]bool{}
	}
	return *u.Permissions
}

// GetPlan returns the Plan field.
//func (u *User) GetPlan() *Plan {
//	if u == nil {
//		return nil
//	}
//	return u.Plan
//}

// GetPrivateGists returns the PrivateGists field if it's non-nil, zero value otherwise.
func (u *User) GetPrivateGists() int {
	if u == nil || u.PrivateGists == nil {
		return 0
	}
	return *u.PrivateGists
}

// GetPublicGists returns the PublicGists field if it's non-nil, zero value otherwise.
func (u *User) GetPublicGists() int {
	if u == nil || u.PublicGists == nil {
		return 0
	}
	return *u.PublicGists
}

// GetPublicRepos returns the PublicRepos field if it's non-nil, zero value otherwise.
func (u *User) GetPublicRepos() int {
	if u == nil || u.PublicRepos == nil {
		return 0
	}
	return *u.PublicRepos
}

// GetReceivedEventsURL returns the ReceivedEventsURL field if it's non-nil, zero value otherwise.
func (u *User) GetReceivedEventsURL() string {
	if u == nil || u.ReceivedEventsURL == nil {
		return ""
	}
	return *u.ReceivedEventsURL
}

// GetReposURL returns the ReposURL field if it's non-nil, zero value otherwise.
func (u *User) GetReposURL() string {
	if u == nil || u.ReposURL == nil {
		return ""
	}
	return *u.ReposURL
}

// GetSiteAdmin returns the SiteAdmin field if it's non-nil, zero value otherwise.
func (u *User) GetSiteAdmin() bool {
	if u == nil || u.SiteAdmin == nil {
		return false
	}
	return *u.SiteAdmin
}

// GetStarredURL returns the StarredURL field if it's non-nil, zero value otherwise.
func (u *User) GetStarredURL() string {
	if u == nil || u.StarredURL == nil {
		return ""
	}
	return *u.StarredURL
}

// GetSubscriptionsURL returns the SubscriptionsURL field if it's non-nil, zero value otherwise.
func (u *User) GetSubscriptionsURL() string {
	if u == nil || u.SubscriptionsURL == nil {
		return ""
	}
	return *u.SubscriptionsURL
}

// GetSuspendedAt returns the SuspendedAt field if it's non-nil, zero value otherwise.
func (u *User) GetSuspendedAt() Timestamp {
	if u == nil || u.SuspendedAt == nil {
		return Timestamp{}
	}
	return *u.SuspendedAt
}

// GetTotalPrivateRepos returns the TotalPrivateRepos field if it's non-nil, zero value otherwise.
func (u *User) GetTotalPrivateRepos() int {
	if u == nil || u.TotalPrivateRepos == nil {
		return 0
	}
	return *u.TotalPrivateRepos
}

// GetTwitterUsername returns the TwitterUsername field if it's non-nil, zero value otherwise.
//func (u *User) GetTwitterUsername() string {
//	if u == nil || u.TwitterUsername == nil {
//		return ""
//	}
//	return *u.TwitterUsername
//}

// GetTwoFactorAuthentication returns the TwoFactorAuthentication field if it's non-nil, zero value otherwise.
func (u *User) GetTwoFactorAuthentication() bool {
	if u == nil || u.TwoFactorAuthentication == nil {
		return false
	}
	return *u.TwoFactorAuthentication
}

// GetType returns the Type field if it's non-nil, zero value otherwise.
func (u *User) GetType() string {
	if u == nil || u.Type == nil {
		return ""
	}
	return *u.Type
}

// GetUpdatedAt returns the UpdatedAt field if it's non-nil, zero value otherwise.
func (u *User) GetUpdatedAt() Timestamp {
	if u == nil || u.UpdatedAt == nil {
		return Timestamp{}
	}
	return *u.UpdatedAt
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (u *User) GetURL() string {
	if u == nil || u.URL == nil {
		return ""
	}
	return *u.URL
}
