package shared

import "github.com/bwmarrin/snowflake"

// Account represents an user account
type Account struct {
	ID       snowflake.ID `json:"id"`
	Username string       `json:"username"`
	Password string       `json:"password,omitempty"`
	Admin    bool         `json:"admin"`
	Created  int64        `json:"created"`
}

// AccountService represents a service which keeps track of user accounts
type AccountService interface {
	Count() (int, error)
	Accounts(skip, limit int) ([]*Account, error)
	Account(id snowflake.ID) (*Account, error)
	AccountByUsername(username string) (*Account, error)
	CreateOrReplace(account *Account) error
	Delete(id snowflake.ID) error
}
