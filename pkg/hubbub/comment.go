// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hubbub

import (
	"time"

	"github.com/google/go-github/v31/github"
)

// GitHubComment is an interface that matches both GitHub issues and pull review comments
type GitHubComment interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *github.Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *github.User
	String() string
}

// Comment is a loose internal comment structure matching issue, PR, code review comments
type Comment struct {
	User        *github.User
	Created     time.Time
	Updated     time.Time
	Body        string
	AuthorAssoc string
	Reactions   *github.Reactions
	ReviewID    int64
}

func NewComment(g GitHubComment) *Comment {
	return &Comment{
		User:        g.GetUser(),
		Body:        g.GetBody(),
		AuthorAssoc: g.GetAuthorAssociation(),
		Created:     g.GetCreatedAt(),
		Updated:     g.GetUpdatedAt(),
		Reactions:   g.GetReactions(),
	}
}
