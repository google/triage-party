package provider

type Reactions struct {
	TotalCount *int    `json:"total_count,omitempty"`
	PlusOne    *int    `json:"+1,omitempty"`
	MinusOne   *int    `json:"-1,omitempty"`
	Laugh      *int    `json:"laugh,omitempty"`
	Confused   *int    `json:"confused,omitempty"`
	Heart      *int    `json:"heart,omitempty"`
	Hooray     *int    `json:"hooray,omitempty"`
	URL        *string `json:"url,omitempty"`
}

// GetConfused returns the Confused field if it's non-nil, zero value otherwise.
func (r *Reactions) GetConfused() int {
	if r == nil || r.Confused == nil {
		return 0
	}
	return *r.Confused
}

// GetHeart returns the Heart field if it's non-nil, zero value otherwise.
func (r *Reactions) GetHeart() int {
	if r == nil || r.Heart == nil {
		return 0
	}
	return *r.Heart
}

// GetHooray returns the Hooray field if it's non-nil, zero value otherwise.
func (r *Reactions) GetHooray() int {
	if r == nil || r.Hooray == nil {
		return 0
	}
	return *r.Hooray
}

// GetLaugh returns the Laugh field if it's non-nil, zero value otherwise.
func (r *Reactions) GetLaugh() int {
	if r == nil || r.Laugh == nil {
		return 0
	}
	return *r.Laugh
}

// GetMinusOne returns the MinusOne field if it's non-nil, zero value otherwise.
func (r *Reactions) GetMinusOne() int {
	if r == nil || r.MinusOne == nil {
		return 0
	}
	return *r.MinusOne
}

// GetPlusOne returns the PlusOne field if it's non-nil, zero value otherwise.
func (r *Reactions) GetPlusOne() int {
	if r == nil || r.PlusOne == nil {
		return 0
	}
	return *r.PlusOne
}

// GetTotalCount returns the TotalCount field if it's non-nil, zero value otherwise.
func (r *Reactions) GetTotalCount() int {
	if r == nil || r.TotalCount == nil {
		return 0
	}
	return *r.TotalCount
}
