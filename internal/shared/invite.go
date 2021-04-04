package shared

// Invite represents a simple invite code
type Invite struct {
	Code    string `json:"code"`
	Created int64  `json:"created"`
}

// InviteService represents the service which keeps track of invite codes
type InviteService interface {
	Count() (int, error)
	Invites(skip, limit int) ([]*Invite, error)
	Invite(code string) (*Invite, error)
	CreateOrReplace(invite *Invite) error
	Delete(code string) error
}
