package daltest

import (
	"testing"

	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/assert"
)

// InitTestDB initializes an in-memory database for testing purposes.
func InitTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
