package migrate

import (
	"bytes"
	"os"
	"testing"

	"github.com/spidernest-go/db/lib/sqlbuilder"
	"github.com/spidernest-go/db/mysql"
	"github.com/stretchr/testify/assert"
)

var Builder sqlbuilder.Database
var Settings = mysql.ConnectionURL{
	Database: "migrate_test",
	Host:     os.Getenv("MYSQL_HOST"),
	User:     os.Getenv("MYSQL_USER"),
	Password: os.Getenv("MYSQL_PASS"),
}

var MigrationName = []string{
	"Creating the users table.",
	"Root admin created.",
	"Add column 'admin' which determines administrative status.",
	"Give 'Root' admin status.",
}
var GoodEntry = []*bytes.Buffer{
	bytes.NewBufferString("CREATE TABLE users(username TEXT NOT NULL)"),
	bytes.NewBufferString("INSERT INTO users(username) VALUES (\"root\")"),
	bytes.NewBufferString("ALTER TABLE users ADD admin BOOL NOT NULL DEFAULT 0"),
	bytes.NewBufferString("UPDATE users SET admin=1 WHERE username=\"root\""),
}

var BadEntry = []*bytes.Buffer{
	bytes.NewBufferString("just a bad sql statement :)"),
}

func TestMain(m *testing.M) {
	var err error
	Builder, err = mysql.Open(Settings)
	if err != nil {
		panic(err)
	}
	defer Builder.Close()
	os.Exit(m.Run())
}

func clear() {
	Builder.Exec("DROP TABLE __meta")
	Builder.Exec("DROP TABLE users")
}

func TestApply(t *testing.T) {
	clear()
	assert.NoError(t, Apply(0, "some name", TestEntry, Builder))
}

func TestLast(t *testing.T) {
	clear()
	migrationName := "some name"
	migrationVer := uint8(0)

	lastMigration, err := Last(Builder)
	assert.Nil(t, lastMigration, "If there are no migrations the result should be nil for the returned migration.")
	assert.NoError(t, err, "No error should occur when the table is queryed for a migration if none exist.")
	assert.NoError(t, Apply(migrationVer, migrationName, TestEntry, Builder), "We expected apply to work but it did not.")

	lastMigration, err = Last(Builder)
	if assert.NotNil(t, lastMigration, "We expected that a migration would be return and it did not get returned.") {
		assert.Equal(t, lastMigration.Name, migrationName, "We expected the migration name in the meta table to be the same as the one we tried to apply.")
		assert.Equal(t, lastMigration.Name, migrationVer, "We expected the migration version in the meta table to be the same as the one we tried to apply.")
	}
	assert.NoError(t, err, "We expected no error retrieving the last migration and we expected one to exist since we just added one.")
}
