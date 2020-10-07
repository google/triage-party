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
	"github.com/google/triage-party/pkg/provider"
)

const (
	reactThumbsUp   = "thumbs_up"
	reactThumbsDown = "thumbs_down"
	reactLaugh      = "laugh"
	reactConfused   = "confused"
	reactHeart      = "heart"
	reactHooray     = "hooray"
)

func reactions(r *provider.Reactions) map[string]int {
	return map[string]int{
		reactThumbsUp:   r.GetPlusOne(),
		reactThumbsDown: r.GetMinusOne(),
		reactLaugh:      r.GetLaugh(),
		reactConfused:   r.GetConfused(),
		reactHeart:      r.GetHeart(),
		reactHooray:     r.GetHooray(),
	}
}
