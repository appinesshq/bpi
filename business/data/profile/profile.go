// Package profile provides CRUD access to the database.
package profile

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotExists = errors.New("profile does not exist")
	ErrExists    = errors.New("profile exists")
	ErrNotFound  = errors.New("profile not found")
)

// Add adds a new profile to the database. If the profile already exists
// this function will fail but the found profile is returned. If the profile is
// being added, the profile with the id from the database is returned.
func Add(ctx context.Context, gql *graphql.GraphQL, n NewProfile) (Profile, error) {
	o := Profile{
		Handle:     n.Handle,
		ScreenName: n.ScreenName,
	}

	o, err := add(ctx, gql, o, n.UserID)
	if err != nil {
		return Profile{}, errors.Wrap(err, "adding profile to database")
	}

	return o, nil
}

// One returns the specified profile from the database by the profile id.
func One(ctx context.Context, gql *graphql.GraphQL, id string) (Profile, error) {
	query := fmt.Sprintf(`
	query {
		getProfile(id: %q) {
			id
			handle
			screen_name
		}
	}`, id)

	var result struct {
		GetProfile Profile `json:"getProfile"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return Profile{}, errors.Wrap(err, "query failed")
	}

	if result.GetProfile.ID == "" {
		return Profile{}, ErrNotFound
	}

	return result.GetProfile, nil
}

// OneByHandle returns the specified profile from the database by handle.
func OneByHandle(ctx context.Context, gql *graphql.GraphQL, handle string) (Profile, error) {
	query := fmt.Sprintf(`
query {
	queryProfile(filter: { handle: { eq: %q } }) {
		id
		handle
		screen_name
	}
}`, handle)

	var result struct {
		QueryProfile []Profile `json:"queryProfile"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return Profile{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryProfile) != 1 {
		return Profile{}, ErrNotFound
	}

	return result.QueryProfile[0], nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, n Profile, userID string) (Profile, error) {
	mutation, result := prepareAdd(n, userID)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return Profile{}, errors.Wrap(err, "failed to add profile")
	}

	if len(result.AddProfile.Profile) != 1 {
		return Profile{}, errors.New("profile id not returned")
	}

	n.ID = result.AddProfile.Profile[0].ID
	return n, nil
}

func prepareAdd(n Profile, userID string) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
mutation {
	addProfile(input: [{
		handle: %q
		screen_name: %q
		user: {
			id: %q
		}
	}])
	%s
}`, n.Handle, n.ScreenName, userID, result.document())

	return mutation, result
}

/*
mutation {
	addProfile(input: [{
		source_id: "1111111111"
    	source: "source"
		screen_name: "goinggodotnet"
		name: "bill kennedy"
		location: "Miami, FL"
	}])
	{
		user {
			id
		}
	}
}

mutation {
	updateProfile(input: {
		filter: {
			id: ["0x04"]
		},
		set: {
			friends: [{
				id: "0x06"
			}]
		}
	})
	{
		numUids
	}
}

mutation {
  updateProfile(input: {
		filter: {
    		id: ["0x04"]
    	},
    	set: {
			friends: [{
				source_id: "4444444444"
				source: "source"
				screen_name: "jacksmith"
				name: "jack smith"
				location: "Miami, FL"
			}]
    	}
  	})
	{
    	numUids
  	}
}

query {
	queryProfile(filter: { screen_name: { eq: "goinggodotnet" } })
	{
		id
		source_id
		source
		screen_name
		name
		location
		friends_count
  	}
}

query {
	getProfile(id: "0x3")
	{
		id
		source_id
		source
		screen_name
		name
		location
		friends_count
	}
}
*/
