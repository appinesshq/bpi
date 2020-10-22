package jurisdiction

type Info struct {
	ID          int    `db:"jurisdiction_id" json:"id"`
	Code        string `db:"code" json:"code"`
	CountryCode string `db:"country_code" json:"country_code"`
	Name        string `db:"name" json:"name"`
	Active      bool   `db:"active" json:"active"`
}
