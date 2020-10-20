package user

import "time"

// User represents someone with access to the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"password,omitempty"`
	Role         string    `json:"role"`
	Profile      Profile   `json:"profile,omitempty"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
}

// Profile is used to capture the user's profile id in relationships.
type Profile struct {
	ID string `json:"id"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
	Role            string `json:"role"`
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
