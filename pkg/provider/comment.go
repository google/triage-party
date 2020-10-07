// Copyright 2020 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"time"
)

// IComment is an interface that matches both GitHub issues and pull review comments
type IComment interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *User
	String() string
}

// Comment is a loose internal comment structure matching issue, PR, code review comments
type Comment struct {
	User        *User
	Created     time.Time
	Updated     time.Time
	Body        string
	AuthorAssoc string
	Reactions   *Reactions
	ReviewID    int64
}

func NewComment(g IComment) *Comment {
	return &Comment{
		User:        g.GetUser(),
		Body:        g.GetBody(),
		AuthorAssoc: g.GetAuthorAssociation(),
		Created:     g.GetCreatedAt(),
		Updated:     g.GetUpdatedAt(),
		Reactions:   g.GetReactions(),
	}
}

func (i *Comment) GetReactions() *Reactions {
	if i == nil {
		return nil
	}
	return i.Reactions
}

func (i *Comment) GetUpdatedAt() time.Time {
	if i == nil {
		return time.Now()
	}
	return i.Updated
}

func (i *Comment) GetCreatedAt() time.Time {
	if i == nil {
		return time.Now()
	}
	return i.Created
}

func (i *Comment) GetAuthorAssociation() string {
	if i == nil {
		return ""
	}
	return i.AuthorAssoc
}

func (i *Comment) GetBody() string {
	if i == nil {
		return ""
	}
	return i.Body
}

func (i *Comment) GetUser() *User {
	if i == nil {
		return nil
	}
	return i.User
}
