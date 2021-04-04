package shared

// Account represents an user account
type Account struct {
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	Admin      bool   `json:"admin"`
	Created    int64  `json:"created"`
	TokenReset int64  `json:"token_reset"`
}

// AccountService represents a service which keeps track of user accounts
type AccountService interface {
	Count() (int, error)
	Accounts(skip, limit int) ([]*Account, error)
	Account(username string) (*Account, error)
	CreateOrReplace(account *Account) error
	Delete(username string) error
}
