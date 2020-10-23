package commands

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mkrou/geonames"
	"github.com/mkrou/geonames/models"
)

func getCountryCode(divCode string) (string, error) {
	p := strings.Split(divCode, ".")
	if len(p) != 2 {
		return "", fmt.Errorf("%q contains no country code", divCode)
	}
	return p[0], nil
}

func JurisdictionSeed() (string, error) {
	b := bytes.Buffer{}
	fmt.Fprint(&b, "INSERT INTO jurisdictions (gnid, code, country_code, name) VALUES")

	p := geonames.NewParser()

	if err := p.GetAdminDivisions(func(d *models.AdminDivision) error {
		countryCode, err := getCountryCode(d.Code)
		if err != nil {
			return err
		}

		fmt.Fprintf(&b, "\n(%d, '%s', '%s', '%s'),",
			d.GeonameId,
			d.Code,
			countryCode,
			strings.Replace(d.Name, "'", "''", -1),
		)
		return nil
	}); err != nil {
		return "", err
	}

	bStr := b.String()
	return fmt.Sprintf("%s\nON CONFLICT DO NOTHING;", bStr[:len(bStr)-1]), nil
}
