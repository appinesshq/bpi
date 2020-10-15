package profile

// Profile represents someone with access to the system.
type Profile struct {
	ID         string `json:"id"`
	Handle     string `json:"handle"`
	ScreenName string `json:"screen_name"`
}

// NewProfile contains information needed to create a new Profile.
type NewProfile struct {
	Handle     string `json:"handle"`
	ScreenName string `json:"screen_name"`
	UserID     string `json:"user_id"`
}

type addResult struct {
	AddProfile struct {
		Profile []struct {
			ID string `json:"id"`
		} `json:"profile"`
	} `json:"addProfile"`
}

func (addResult) document() string {
	return `{
		profile {
			id
		}
	}`
}
