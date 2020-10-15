package user

// User represents someone with access to the system.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"location,omitempty"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Email    string `json:"email"`
	Password string `json:"location"`
}

type addResult struct {
	AddUser struct {
		User []struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"addUser"`
}

func (addResult) document() string {
	return `{
		user {
			id
		}
	}`
}
