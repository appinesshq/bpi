package jurisdiction

type Info struct {
	Code        string `db:"code" json:"code"`
	GNID        int    `db:"gnid" json:"gnid"`
	CountryCode string `db:"country_code" json:"country_code"`
	Name        string `db:"name" json:"name"`
	Active      bool   `db:"active" json:"active"`
}
