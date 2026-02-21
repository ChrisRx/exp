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

	t.Run("", func(t *testing.T) {
		s := Dedent(`Switched from using alpine.js to using htmx and changed the Netscape
			browser buttons to navigate to different pages. Seems pretty neat so far!

			Some other paragraph [a link](https://www.google.com)`)

		assert.Equal(t, `Switched from using alpine.js to using htmx and changed the Netscape
browser buttons to navigate to different pages. Seems pretty neat so far!

Some other paragraph [a link](https://www.google.com)
`, s)
	})
}

func TestSlug(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{input: "some_test.MyStruct", expected: "some-test-mystruct"},
		{input: "The Séance of Blake Manor", expected: "the-seance-of-blake-manor"},
		{input: "", expected: ""},
		{input: "This & That!", expected: "this-and-that"},
		{input: "à", expected: "a"},
	}

	for _, tt := range cases {
		assert.Equal(t, tt.expected, Slug(tt.input), tt.name)
	}
}

func TestCamelCase(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{input: "name", expected: "Name"},
		{input: "name_", expected: "Name"},
		{input: "first_name", expected: "FirstName"},
		{input: "_firsT_Name_", expected: "FirstName"},
		{input: "first__name", expected: "FirstName"},
		{input: "", expected: ""},
		{input: "	", expected: ""},
		{input: "    ", expected: ""},
	}

	for _, tt := range cases {
		assert.Equal(t, tt.expected, ToCamelCase(tt.input), tt.name)
	}
}
