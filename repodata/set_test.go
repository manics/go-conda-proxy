package repodata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAddRemove(t *testing.T) {
	s := NewSet(nil)
	assert.Equal(t, 0, s.Len(), "Expected empty set")
	assert.Equal(t, []string{}, *s.Items(), "Expected 0 items")

	s.Add("foo")
	assert.Equal(t, 1, s.Len())
	assert.Equal(t, []string{"foo"}, *s.Items())
	assert.True(t, s.Contains("foo"))
	assert.False(t, s.Contains("bar"))

	s.Add("bar")
	assert.Equal(t, 2, s.Len())
	assert.ElementsMatch(t, []string{"bar", "foo"}, *s.Items())
	assert.True(t, s.Contains("foo"))
	assert.True(t, s.Contains("bar"))

	s.Remove("foo")
	assert.Equal(t, 1, s.Len())
	assert.Equal(t, []string{"bar"}, *s.Items())
	assert.False(t, s.Contains("foo"))
	assert.True(t, s.Contains("bar"))

	// Doesn't throw error (should it?)
	s.Remove("foo")
	assert.Equal(t, []string{"bar"}, *s.Items())
}

func TestSetInitPop(t *testing.T) {
	s := NewSet(&[]string{"foo", "bar"})

	assert.Equal(t, 2, s.Len())
	assert.ElementsMatch(t, []string{"bar", "foo"}, *s.Items())
	assert.True(t, s.Contains("foo"))
	assert.True(t, s.Contains("bar"))

	a, valid := s.Pop()
	assert.True(t, valid)
	b, valid := s.Pop()
	assert.True(t, valid)
	assert.ElementsMatch(t, []string{"foo", "bar"}, []string{a, b})

	_, valid = s.Pop()
	assert.False(t, valid)
}
