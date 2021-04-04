package shared

// Mailbox represents a simple mailbox mapped to an user account
type Mailbox struct {
	Address string `json:"address"`
	Account string `json:"account"`
}

// MailboxService represents a service which keeps track of mailboxes
type MailboxService interface {
	Count() (int, error)
	Mailboxes(skip, limit int) ([]*Mailbox, error)
	CountWithAccount(account string) (int, error)
	MailboxesWithAccount(account string) ([]*Mailbox, error)
	Mailbox(address string) (*Mailbox, error)
	CreateOrReplace(mailbox *Mailbox) error
	Transfer(oldAccount, newAccount string) error
	Delete(address string) error
	DeleteWithAccount(account string) error
}
