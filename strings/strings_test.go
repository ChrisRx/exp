package strings

import (
	"testing"

	"go.chrisrx.dev/x/assert"
)

func TestDedent(t *testing.T) {
	t.Run("", func(t *testing.T) {
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
	})

	t.Run("", func(t *testing.T) {
		s := Dedent(`
		Switched from using alpine.js to using htmx and changed the Netscape
		browser buttons to navigate to different pages. Seems pretty neat so far!

		Some other paragraph [a link](https://www.google.com)
	`)

		assert.Equal(t, `Switched from using alpine.js to using htmx and changed the Netscape
browser buttons to navigate to different pages. Seems pretty neat so far!

Some other paragraph [a link](https://www.google.com)
`, s)
	})

}

func TestSlug(t *testing.T) {
	s := Slug("some_test.MyStruct")
	assert.Equal(t, "some-test-mystruct", s)
}
