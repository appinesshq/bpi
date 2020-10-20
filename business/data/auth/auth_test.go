package auth_test

import (
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/appinesshq/bpi/business/data/auth"
	"github.com/appinesshq/bpi/foundation/tests"
	"github.com/dgrijalva/jwt-go"
)

func TestAuthenticator(t *testing.T) {
	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateRSAKey))
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the private key from pem: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the private key from pem.", tests.Success, testID)

			// The key id we are stating represents the public key in the
			// public key store.
			const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

			keyLookupFunc := func(publicKID string) (*rsa.PublicKey, error) {
				if publicKID != keyID {
					return nil, errors.New("no public key found")
				}
				return &privateKey.PublicKey, nil
			}
			a, err := auth.New(privateKey, keyID, "RS256", keyLookupFunc)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", tests.Success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "travel project",
					Subject:   "0x01",
					Audience:  "students",
					ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
					IssuedAt:  time.Now().Unix(),
				},
				Auth: auth.StandardClaims{
					Role: auth.RoleAdmin,
				},
			}

			token, err := a.GenerateToken(claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", tests.Success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", tests.Success, testID)

			if exp, got := claims.Auth.Role, parsedClaims.Auth.Role; exp != got {
				t.Logf("\t\tTest %d:\texp: %v", testID, exp)
				t.Logf("\t\tTest %d:\tgot: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expexted roles: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expexted roles.", tests.Success, testID)
		}
	}
}

// Output of:
// openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
// ./bpi-admin keygen
const privateRSAKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0cyRhWgtqIAbL1FAJZSVi3azSPsyYel7JBjQez1Uoa+hXjlA
AYum+JQE3GaTnCwseT8r9NrC+Oy3/1S6TIbhLfZuzX13ZFDQwPsZAJOKWDDLbgiZ
Benls4bV8fn+o+MFzMUtI0iNuqXvfLc6okFDq01dKktNyV43zWvY8qd4rZfcOiSj
XvhbY+Ph5MXtg+pc2u/rOdZAM5FrxM5J8dHgaNg/DkUyN1KTGTB0UesJ6CL52SpB
KJG4IWNU257DrePwq2LoETQBfxcWiYy9Yojr0VEGPfp5QL8YCIQVxDmdAhlHOsfs
crdIEGutS48AyIcarnY2TbSfgzkQX48zYur0WwIDAQABAoIBAQCkbMKEHtDh5Xzo
ybIPgfLuOZprkUu8RwOWl8gVPkzs5zv+H7pVO8EhwshIgDAhztEQOX1WynjSJJxU
BXB496DVp/TRIgsHWPsys9i1heyAD8Xvt9dONjErUXqtybNTeGKcSNCGfZ9ucAxQ
3z2Z2rKRN/HTau9M6YWsfmCqVKyUxyt1I+J2WUpApQjsCW+LXo8QWg9dciUfq2bM
ygOKNsKeaB7Yd93Jo7XOG1CHadQnnxIFa18wknoJGNznNBsWmtkTnV5RMDLTzs4V
LKIYseNta+T44rNJ+ymZ1N6K7CKi9na6iodBJhMplynfnHCDKGzZz9+G2o36JizI
E3CIjiEZAoGBAN9j9v8/HIa29wW87H+Egx7aIogfTu3CDvnMPLQMItee9hdPvj2d
/plCiAqkxzb0HFMhGK3XY2lxIua4gL3HKc5op/B4agzNNIW8KatLNoV1gdS/HybX
6Eznds+3BzZLYoXGVTcpeWIOXW2AYbEKOFkazC4uFsctoUb5igPUlJXtAoGBAPBs
soewVcJ0LfChT5Sj/69JtaYsLgsdzGW9LHiGSnmtqiCaogSdjA9mmVj7GFAWbOM2
xz56PKI8ESuxgQGn/0KmderW9BJZtSQLK/WzKSuk3XsbdvxHwxfhOknYc4+c73oF
swoT6PE/uQNjGMa4vFZWR+6uTUrXXnbswVVveOpnAoGAFNEA5DoiU19bV3qKYzua
6FYVX6/jL/6kXJyuj2yOFp+mePeiV6WQYwGzIaLHOZS3yvtLjG+EwP+c6/kHbifP
+n8AH0VpRHYezdOB4odotjkD9yo0Ie9+oyPyi1qX3nRZ4vNfX3uK5xtFk32iHNhB
9fOsUSVUVA0peS6psL+vdOUCgYAXgvm+jT8FwijP9GZ86cDSWon6Ey35hlN7y5Ey
xCc6WQJfJ+AaRXHx+52Zdwy8oETLv4qikH+neepP9I7iI5Sx5ud3LMg3lzBAsxr8
byXij7/dDyWGrFnm1u7FU/aRH87HhxEoNiQ8m3ezXhiJLn20j8F/FOqYHBGv3Z1W
ho0zlwKBgB/wxJ/c5vgF0Y+MlTHFK8QaMSpHUV0lIey/WLfb6kC4MrL0boOg+nAh
hZEJyVcCZU0NKYQQYsbpMZDfka1kojmC9CeQMaMh52lfIsNJJxvLyHukcdInE0mB
XogQ/HIj3XDhm7vcY66oeRedbF+iSWqUOYqCOvjNMqzFPEu7p1Im
-----END RSA PRIVATE KEY-----`

// To generate a public key PEM file.
// openssl rsa -pubout -in private.pem -out public.pem
// ./bpi-admin keygen
