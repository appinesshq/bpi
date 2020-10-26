// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sqlx.DB) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)
	return d.Migrate()
}

// migrations contains the queries needed to construct the database schema.
// Entries should never be removed once they have been run in production.
//
// Using constants in a .go file is an easy way to ensure the schema is part
// of the compiled executable and avoids pathing issues with the working
// directory. It has the downside that it lacks syntax highlighting and may be
// harder to read for some cases compared to using .sql files. You may also
// consider a combined approach using a tool like packr or go-bindata.
var migrations = []darwin.Migration{
	{
		Version:     1.1,
		Description: "Create table users",
		Script: `
CREATE TABLE users (
	user_id       UUID,
	email         TEXT UNIQUE,
	roles         TEXT[],
	password_hash TEXT,
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,

	PRIMARY KEY (user_id)
);`,
	},
	{
		Version:     1.2,
		Description: "Create table profiles",
		Script: `
CREATE TABLE profiles (
	name     TEXT UNIQUE,
	display_name  TEXT,
	type 		  TEXT,
	user_id UUID  DEFAULT '00000000-0000-0000-0000-000000000000',
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,

	PRIMARY KEY (name)
);`,
	},
	{
		Version:     1.3,
		Description: "Create table countries",
		Script: `
CREATE TABLE countries (
	code  		  TEXT UNIQUE,
	gnid    	  INT UNIQUE,
	name 		  TEXT,
	currency_code TEXT,
	currency_name TEXT,
	active BOOL DEFAULT FALSE,

	PRIMARY KEY (code)
);`,
	},
	{
		Version:     1.4,
		Description: "Create table jurisdiction",
		Script: `
CREATE TABLE jurisdictions (
	code			TEXT UNIQUE,
	gnid 			INT UNIQUE,
	country_code 	TEXT,
	name		 	TEXT,
	active 			BOOL DEFAULT FALSE,

	PRIMARY KEY (code),
	FOREIGN KEY (country_code) REFERENCES countries(code) ON DELETE CASCADE
);`,
	},
	// FOREIGN KEY (parent_id) REFERENCES categories(category_id) ON DELETE SET NULL
	{
		Version:     1.5,
		Description: "Create table categories",
		Script: `
CREATE TABLE categories (
	category_id   UUID,
	slug          TEXT UNIQUE,
	name 		  TEXT,
	user_id 	  UUID  DEFAULT '00000000-0000-0000-0000-000000000000',
	parent_id     UUID  REFERENCES categories(category_id),
	date_created TIMESTAMP,
	date_updated TIMESTAMP,
	
	PRIMARY KEY (category_id)
);`,
	},
	{
		Version:     2.1,
		Description: "Create table products",
		Script: `
CREATE TABLE products (
	product_id   UUID,
	name         TEXT,
	cost         INT,
	quantity     INT,
	user_id UUID DEFAULT '00000000-0000-0000-0000-000000000000',
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (product_id)
);`,
	},
	{
		Version:     2.2,
		Description: "Create table sales",
		Script: `
CREATE TABLE sales (
	sale_id      UUID,
	product_id   UUID,
	quantity     INT,
	paid         INT,
	date_created TIMESTAMP,

	PRIMARY KEY (sale_id),
	FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
);`,
	},
}
