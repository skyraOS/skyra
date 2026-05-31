package reality

type User struct {
	id      string
	BeingID string
	Content string
	Weight  float64
	Type    string

	ObserveFns map[string]func(*Relation)
	ExpressFns map[string]func(*Relation) *Relation
}

func NewUser(beingID, content string) *User {
	return &User{
		id:         beingID,
		BeingID:    beingID,
		Content:    content,
		Weight:     1.0,
		Type:       "terminal",
		ObserveFns: make(map[string]func(*Relation)),
		ExpressFns: make(map[string]func(*Relation) *Relation),
	}
}

func (u *User) ID() string { return u.id }

func (u *User) Create(r *Relation) Reality { return nil }

func (u *User) Realize(r *Relation) *Relation {
	u.Competence(r)
	u.Observe(r)
	r = u.Express(r)
	return r
}

// Competence — selects the device. Deterministic for now.
func (u *User) Competence(r *Relation) {
	u.Type = "terminal"
}

// Observe — receive from the world through the selected device
func (u *User) Observe(r *Relation) {
	if fn, ok := u.ObserveFns[u.Type]; ok {
		fn(r)
	}
}

// Express — emit into the world through the selected device
func (u *User) Express(r *Relation) *Relation {
	if fn, ok := u.ExpressFns[u.Type]; ok {
		r = fn(r)
	}
	return r
}

func (u *User) Activation(r *Relation) float64 {
	return u.Weight
}

func (u *User) Reinforce() {}
func (u *User) Decay()     {}
