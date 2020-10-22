package commands

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mkrou/geonames"
	"github.com/mkrou/geonames/models"
)

func CountrySeed() (string, error) {
	b := bytes.Buffer{}
	fmt.Fprint(&b, "INSERT INTO countries (country_id, country_code, name, currency_code, currency_name) VALUES")

	p := geonames.NewParser()

	if err := p.GetCountries(func(country *models.Country) error {
		fmt.Fprintf(&b, "\n(%d, '%s', '%s', '%s', '%s'),",
			country.GeonameID,
			strings.Replace(country.Iso2Code, "'", "''", -1),
			strings.Replace(country.Name, "'", "''", -1),
			strings.Replace(country.CurrencyCode, "'", "''", -1),
			strings.Replace(country.CurrencyName, "'", "''", -1),
		)
		return nil
	}); err != nil {
		return "", err
	}

	bStr := b.String()
	return fmt.Sprintf("%s\nON CONFLICT DO NOTHING;", bStr[:len(bStr)-1]), nil

	// Print all cities with a population greater than 500
	// err := p.GetGeonames(geonames.Cities500, func(geoname *models.Geoname) error {
	// 	fmt.Println(geoname.Name)
	// 	return nil
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
