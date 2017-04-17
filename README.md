# Tachyon
#### Tachyon is a simple Fixture library for you Go test suite.

Tachyon loads your Yaml fixtures into your database in an unopinionated manner.

Tachyon is still a work in progress, so expect great changes to come!

## Installation

The easy way:
```sh
go get github.com/ki4jnq/tachyon
```

However, it's recommended that you use a tool such as `glide` and vendor your dependencies.

## Usage

```go
import "github.com/ki4jnq/tachyon"

func TestSomethingBig(t *testing.T) {
	var db *sql.Db
	db := magicallyMakeDbInstance()

	// Read one fixture from testdata/users.yml
	f, err := tachyon.NewFixture("users")

	// Load the fixture into the database.
	err = f.Load(db)

	// ... Or, load it in a transaction
	tx, _ := db.Begin()
	err = f.LoadTx(tx)
	tx.Commit() // It's left up to you manage the transaction.

	// ... Or manage multiple fixtures at once
	fixtures, err := tachyon.ReadFixtures("users", "posts", "comments")

	// Load them all into the DB
	fixtures.Load(db)

	// And when you're done:
	fixtures.Clean(db)
}
```

Fixture files should be in the following format:

```yaml
table: users
fields:
  - email
  - name

data:
  picard:
    email: "picard.cc@enterprise.net"
    name: "Jean Luc Picard"

  riker:
    email: "riker.xo@enterprise.net"
    name: "William Riker"
```

Place your fixtures in a `testdata/` subdirectory relative to the root package path. Make sure that your fixtures have a `.yml` file extension.

## TODO

Loading a list of fixtures creates 1 transaction per fixture, it should only create a single transaction, period.

## Special Thanks

Special thanks to [Tax Management Associates (TMA)](http://www.tma1.com/) for letting me have fun and call it "work".
