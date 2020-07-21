package interfaces

import (
	"github.com/google/triage-party/pkg/models"
	"time"
)

// GitHubComment is an interface that matches both GitHub issues and pull review comments
type IComment interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *models.Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *models.User
	String() string
}
