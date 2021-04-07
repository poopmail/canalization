package shared

import "github.com/bwmarrin/snowflake"

// Mailbox represents a simple mailbox mapped to an user account
type Mailbox struct {
	Address string       `json:"address"`
	Account snowflake.ID `json:"account"`
	Created int64        `json:"created"`
}

// MailboxService represents a service which keeps track of mailboxes
type MailboxService interface {
	Count() (int, error)
	Mailboxes(skip, limit int) ([]*Mailbox, error)
	CountInAccount(account snowflake.ID) (int, error)
	MailboxesInAccount(account snowflake.ID, skip, limit int) ([]*Mailbox, error)
	Mailbox(address string) (*Mailbox, error)
	CreateOrReplace(mailbox *Mailbox) error
	Delete(address string) error
	DeleteInAccount(account snowflake.ID) error
}
