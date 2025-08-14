package set

import (
	"encoding/json"
	"testing"

	"go.chrisrx.dev/x/assert"
)

func TestSet(t *testing.T) {
	t.Run("equals", func(t *testing.T) {
		var s1, s2 Set[int]
		s1.Add(1)
		s1.Add(2)
		s1.Add(3)

		s2.Add(1)
		s2.Add(2)
		s2.Add(3)
		s2.Add(4)

		assert.Equal(t, false, s1.Equals(&s2))
		s1.Add(4)
		assert.Equal(t, true, s1.Equals(&s2))
		s1.Add(5)
		assert.Equal(t, false, s1.Equals(&s2))
	})

	t.Run("all", func(t *testing.T) {
		var s Set[int]
		s.Add(1)
		s.Add(2)
		s.Add(3)

		assert.Equal(t, false, s.All(1, 2, 4))
		s.Add(4)
		assert.Equal(t, true, s.All(1, 2, 4))
	})

	t.Run("any", func(t *testing.T) {
		var s Set[int]
		s.Add(1)
		s.Add(2)
		s.Add(3)

		assert.Equal(t, false, s.Any(4, 5, 6))
		s.Add(4)
		assert.Equal(t, true, s.Any(4, 5, 6))
	})

	t.Run("union", func(t *testing.T) {
		var s1, s2 Set[int]
		s1.Add(1)
		s1.Add(2)
		s1.Add(4)

		s2.Add(2)
		s2.Add(3)
		s2.Add(5)
		s2.Add(7)

		s3 := s1.Union(&s2)

		assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 7}, s3.List())
	})

	t.Run("difference", func(t *testing.T) {
		var s1, s2 Set[int]
		s1.Add(1)
		s1.Add(2)
		s1.Add(4)

		s2.Add(2)
		s2.Add(3)
		s2.Add(5)
		s2.Add(7)

		s3 := s1.Difference(&s2)

		assert.ElementsMatch(t, []int{1, 4}, s3.List())
	})

	t.Run("intersection", func(t *testing.T) {
		var s1, s2 Set[int]
		s1.Add(1)
		s1.Add(2)
		s1.Add(4)
		s1.Add(7)

		s2.Add(2)
		s2.Add(3)
		s2.Add(5)
		s2.Add(7)

		s3 := s1.Intersection(&s2)

		assert.ElementsMatch(t, []int{2, 7}, s3.List())
	})

	t.Run("set of sets", func(t *testing.T) {
		s := New(
			New(1, 1, 2, 3, 3, 3),
			New(1, 2, 3),
			New(1, 1, 1, 2, 2, 3, 3, 3),
			New(1, 2, 2, 2, 3),
		)

		assert.Equal(t, 1, s.Len())
		assert.ElementsMatch(t, []int{1, 2, 3}, s.Pop().List())
	})

	t.Run("marshal/unmarshal", func(t *testing.T) {
		s1 := New(1, 1, 1, 2, 2, 3, 3, 3)
		data, err := json.Marshal(s1)
		if err != nil {
			t.Fatal(err)
		}

		var s2 Set[int]
		if err := json.Unmarshal(data, &s2); err != nil {
			t.Fatal(err)
		}

		assert.ElementsMatch(t, s1.List(), s2.List())
	})
}

type uncomparable struct {
	data []byte
}

func TestSetUncomparable(t *testing.T) {
	expected := []byte("some value")

	var s Set[uncomparable]
	for range 10 {
		s.Add(uncomparable{data: expected})
	}

	assert.Equal(t, 1, s.Len())
	assert.Equal(t, expected, s.Pop().data)
	assert.Equal(t, 0, s.Len())

	t.Run("union uncomparable", func(t *testing.T) {
		var s1, s2 Set[uncomparable]
		s1.Add(uncomparable{data: []byte("1")})
		s1.Add(uncomparable{data: []byte("2")})
		s1.Add(uncomparable{data: []byte("4")})

		s2.Add(uncomparable{data: []byte("2")})
		s2.Add(uncomparable{data: []byte("3")})
		s2.Add(uncomparable{data: []byte("5")})
		s2.Add(uncomparable{data: []byte("7")})

		s3 := s1.Union(&s2)

		expected := []uncomparable{
			{data: []byte("1")},
			{data: []byte("2")},
			{data: []byte("3")},
			{data: []byte("4")},
			{data: []byte("5")},
			{data: []byte("7")},
		}
		assert.ElementsMatch(t, expected, s3.List())
	})

	t.Run("uncomparable set of sets", func(t *testing.T) {
		s := New(
			New([]byte("1"), []byte("2"), []byte("3")),
			New([]byte("1"), []byte("2"), []byte("3")),
			New([]byte("1"), []byte("2"), []byte("3"), []byte("3")),
			New([]byte("1"), []byte("1"), []byte("2"), []byte("3")),
		)

		assert.Equal(t, 1, s.Len())
		assert.ElementsMatch(t, [][]byte{[]byte("1"), []byte("2"), []byte("3")}, s.Pop().List())
	})
}

func BenchmarkEquals(b *testing.B) {
	s1 := New(1, 2, 3, 4, 5, 6, 7, 8, 9)
	s2 := New(1, 2, 3, 4, 5, 6, 8, 9, 10)
	s3 := New(1, 2, 3, 4, 5, 6, 8, 9)
	b.Run("hashset.all", func(b *testing.B) {
		for b.Loop() {
			s1.All(s2.List()...)
		}
	})
	b.Run("Compare", func(b *testing.B) {
		for b.Loop() {
			Compare(s1, s2)
		}
	})
	b.Run("Compare (length mismatch)", func(b *testing.B) {
		for b.Loop() {
			Compare(s1, s3)
		}
	})
}
