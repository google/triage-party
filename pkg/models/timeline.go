package models

import "time"

// Timeline represents an event that occurred around an Issue or Pull Request.
//
// It is similar to an IssueEvent but may contain more information.
// GitHub API docs: https://developer.github.com/v3/issues/timeline/
type Timeline struct {
	ID        *int64  `json:"id,omitempty"`
	URL       *string `json:"url,omitempty"`
	CommitURL *string `json:"commit_url,omitempty"`

	// The User object that generated the event.
	Actor *User `json:"actor,omitempty"`

	// Event identifies the actual type of Event that occurred. Possible values
	// are:
	//
	//     assigned
	//       The issue was assigned to the assignee.
	//
	//     closed
	//       The issue was closed by the actor. When the commit_id is present, it
	//       identifies the commit that closed the issue using "closes / fixes #NN"
	//       syntax.
	//
	//     commented
	//       A comment was added to the issue.
	//
	//     committed
	//       A commit was added to the pull request's 'HEAD' branch. Only provided
	//       for pull requests.
	//
	//     cross-referenced
	//       The issue was referenced from another issue. The 'source' attribute
	//       contains the 'id', 'actor', and 'url' of the reference's source.
	//
	//     demilestoned
	//       The issue was removed from a milestone.
	//
	//     head_ref_deleted
	//       The pull request's branch was deleted.
	//
	//     head_ref_restored
	//       The pull request's branch was restored.
	//
	//     labeled
	//       A label was added to the issue.
	//
	//     locked
	//       The issue was locked by the actor.
	//
	//     mentioned
	//       The actor was @mentioned in an issue body.
	//
	//     merged
	//       The issue was merged by the actor. The 'commit_id' attribute is the
	//       SHA1 of the HEAD commit that was merged.
	//
	//     milestoned
	//       The issue was added to a milestone.
	//
	//     referenced
	//       The issue was referenced from a commit message. The 'commit_id'
	//       attribute is the commit SHA1 of where that happened.
	//
	//     renamed
	//       The issue title was changed.
	//
	//     reopened
	//       The issue was reopened by the actor.
	//
	//     subscribed
	//       The actor subscribed to receive notifications for an issue.
	//
	//     unassigned
	//       The assignee was unassigned from the issue.
	//
	//     unlabeled
	//       A label was removed from the issue.
	//
	//     unlocked
	//       The issue was unlocked by the actor.
	//
	//     unsubscribed
	//       The actor unsubscribed to stop receiving notifications for an issue.
	//
	Event *string `json:"event,omitempty"`

	// The string SHA of a commit that referenced this Issue or Pull Request.
	CommitID *string `json:"commit_id,omitempty"`
	// The timestamp indicating when the event occurred.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// The Label object including `name` and `color` attributes. Only provided for
	// 'labeled' and 'unlabeled' events.
	Label *Label `json:"label,omitempty"`
	// The User object which was assigned to (or unassigned from) this Issue or
	// Pull Request. Only provided for 'assigned' and 'unassigned' events.
	Assignee *User `json:"assignee,omitempty"`
	// The Milestone object including a 'title' attribute.
	// Only provided for 'milestoned' and 'demilestoned' events.
	Milestone *Milestone `json:"milestone,omitempty"`

	// The 'id', 'actor', and 'url' for the source of a reference from another issue.
	// Only provided for 'cross-referenced' events.
	Source *Source `json:"source,omitempty"`

	// An object containing rename details including 'from' and 'to' attributes.
	// Only provided for 'renamed' events.
	//Rename      *Rename      `json:"rename,omitempty"`

	//ProjectCard *ProjectCard `json:"project_card,omitempty"`
}

// GetActor returns the Actor field.
func (t *Timeline) GetActor() *User {
	if t == nil {
		return nil
	}
	return t.Actor
}

// GetAssignee returns the Assignee field.
func (t *Timeline) GetAssignee() *User {
	if t == nil {
		return nil
	}
	return t.Assignee
}

// GetCommitID returns the CommitID field if it's non-nil, zero value otherwise.
func (t *Timeline) GetCommitID() string {
	if t == nil || t.CommitID == nil {
		return ""
	}
	return *t.CommitID
}

// GetCommitURL returns the CommitURL field if it's non-nil, zero value otherwise.
func (t *Timeline) GetCommitURL() string {
	if t == nil || t.CommitURL == nil {
		return ""
	}
	return *t.CommitURL
}

// GetCreatedAt returns the CreatedAt field if it's non-nil, zero value otherwise.
func (t *Timeline) GetCreatedAt() time.Time {
	if t == nil || t.CreatedAt == nil {
		return time.Time{}
	}
	return *t.CreatedAt
}

// GetEvent returns the Event field if it's non-nil, zero value otherwise.
func (t *Timeline) GetEvent() string {
	if t == nil || t.Event == nil {
		return ""
	}
	return *t.Event
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (t *Timeline) GetID() int64 {
	if t == nil || t.ID == nil {
		return 0
	}
	return *t.ID
}

// GetLabel returns the Label field.
func (t *Timeline) GetLabel() *Label {
	if t == nil {
		return nil
	}
	return t.Label
}

// GetMilestone returns the Milestone field.
func (t *Timeline) GetMilestone() *Milestone {
	if t == nil {
		return nil
	}
	return t.Milestone
}

// GetProjectCard returns the ProjectCard field.
//func (t *Timeline) GetProjectCard() *ProjectCard {
//	if t == nil {
//		return nil
//	}
//	return t.ProjectCard
//}
//
//// GetRename returns the Rename field.
//func (t *Timeline) GetRename() *Rename {
//	if t == nil {
//		return nil
//	}
//	return t.Rename
//}

// GetSource returns the Source field.
func (t *Timeline) GetSource() *Source {
	if t == nil {
		return nil
	}
	return t.Source
}

// GetURL returns the URL field if it's non-nil, zero value otherwise.
func (t *Timeline) GetURL() string {
	if t == nil || t.URL == nil {
		return ""
	}
	return *t.URL
}
