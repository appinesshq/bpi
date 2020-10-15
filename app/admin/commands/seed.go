package commands

import (
	"context"
	"log"
	"time"

	"github.com/appinesshq/bpi/business/data"
)

// Seed will seed the database for a given user.
func Seed(log *log.Logger, gqlConfig data.GraphQLConfig) error {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: Seed db with countries and jurisdictions.

	// t := twitter.New(log, token)
	// u, err := t.RetrieveUser(ctx, screenName)
	// if err != nil {
	// 	return err
	// }

	// _, err = t.RetrieveFriends(ctx, u.ID)
	// if err != nil {
	// 	return err
	// }

	return nil
}
