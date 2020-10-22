package country

type Info struct {
	ID           int    `db:"country_id" json:"id"`
	CountryCode  string `db:"code" json:"code"`
	Name         string `db:"name" json:"name"`
	CurrencyCode string `db:"currency_code" json:"currency_code"`
	CurrencyName string `db:"currency_name" json:"currency_name"`
	Active       bool   `db:"active" json:"active"`
}
