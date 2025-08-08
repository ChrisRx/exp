package strings

import (
	"testing"

	"go.chrisrx.dev/x/assert"
)

func TestDedent(t *testing.T) {
	s := Dedent(`
		SELECT
			*
		FROM
			table t1
		WHERE
			status = 'ok'
			AND something <> 'lol'
	`)

	assert.Equal(t, `SELECT
	*
FROM
	table t1
WHERE
	status = 'ok'
	AND something <> 'lol'
`, s)
}
