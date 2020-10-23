package country

type Info struct {
	CountryCode  string `db:"code" json:"code"`
	GNID         int    `db:"gnid" json:"gnid"`
	Name         string `db:"name" json:"name"`
	CurrencyCode string `db:"currency_code" json:"currency_code"`
	CurrencyName string `db:"currency_name" json:"currency_name"`
	Active       bool   `db:"active" json:"active"`
}
