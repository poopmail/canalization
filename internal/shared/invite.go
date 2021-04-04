package shared

// Invite represents a simple invite code
type Invite string

// InviteService represents the service which keeps track of invite codes
type InviteService interface {
	Count() (int, error)
	Invites(skip, limit int) ([]Invite, error)
	Validate(invite string) (bool, error)
	Create(invite Invite) error
	Delete(invite string) error
}
