// Package user provides CRUD access to the database.
package user

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotExists = errors.New("user does not exist")
	ErrExists    = errors.New("user exists")
	ErrNotFound  = errors.New("user not found")
)

// Add adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func Add(ctx context.Context, gql *graphql.GraphQL, nu NewUser) (User, error) {
	u := User{
		Email:    nu.Email,
		Password: nu.Password,
	}

	u, err := add(ctx, gql, u)
	if err != nil {
		return User{}, errors.Wrap(err, "adding user to database")
	}

	return u, nil
}

// One returns the specified user from the database by the user id.
func One(ctx context.Context, gql *graphql.GraphQL, userID string) (User, error) {
	query := fmt.Sprintf(`
	query {
		getUser(id: %q) {
			id
			email
			profile {
				id
			}
		}
	}`, userID)

	var result struct {
		GetUser User `json:"getUser"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if result.GetUser.ID == "" {
		return User{}, ErrNotFound
	}

	return result.GetUser, nil
}

// OneByEmail returns the specified user from the database by email.
func OneByEmail(ctx context.Context, gql *graphql.GraphQL, email string) (User, error) {
	query := fmt.Sprintf(`
query {
	queryUser(filter: { email: { eq: %q } }) {
		id
		email
		profile {
			id
		}
	}
}`, email)

	var result struct {
		QueryUser []User `json:"queryUser"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryUser) != 1 {
		return User{}, ErrNotFound
	}

	return result.QueryUser[0], nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, user User) (User, error) {
	mutation, result := prepareAdd(user)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return User{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.AddUser.User) != 1 {
		return User{}, errors.New("user id not returned")
	}

	user.ID = result.AddUser.User[0].ID
	return user, nil
}

func prepareAdd(user User) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
mutation {
	addUser(input: [{
		email: %q
		password: %q
	}])
	%s
}`, user.Email, user.Password, result.document())

	return mutation, result
}

/*
mutation {
	addUser(input: [{
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
	updateUser(input: {
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
  updateUser(input: {
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
	queryUser(filter: { screen_name: { eq: "goinggodotnet" } })
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
	getUser(id: "0x3")
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
