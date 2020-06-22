package configstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTreeProvider(t *testing.T) {
	var s = NewStore()
	s.FileTree("tests/fixtures/filetreeprovider")
	l, err := s.GetItemList()

	assert.Equal(t, 6, l.Len())

	barItems, has := l.indexed["bar"]
	require.True(t, has, "missing 'bar' items")
	require.Len(t, barItems, 2, "there must be 2 'bar' items")

	biz := barItems[0]
	require.Equal(t, "bar", biz.Key())
	require.Equal(t, "biz value", biz.value)

	buz := barItems[1]
	require.Equal(t, "bar", buz.Key())
	require.Equal(t, "buz value", buz.value)

	bar_bizItems, has := l.indexed["bar/biz"]
	require.True(t, has, "missing 'bar/biz' items")
	require.Len(t, bar_bizItems, 1, "there must be 1 'bar/biz' item")

	bar_biz := bar_bizItems[0]
	require.Equal(t, "bar/biz", bar_biz.Key())
	require.Equal(t, "biz value", bar_biz.value)

	bar_buzItems, has := l.indexed["bar/buz"]
	require.True(t, has, "missing 'bar/buz' items")
	require.Len(t, bar_buzItems, 1, "there must be 1 'bar/buz' item")

	bar_buz := bar_buzItems[0]
	require.Equal(t, "bar/buz", bar_buz.Key())
	require.Equal(t, "buz value", bar_buz.value)

	bazItems, has := l.indexed["baz"]
	require.True(t, has, "missing 'baz' items")
	require.Len(t, bazItems, 1, "there must be 1 'baz' item")

	baz := bazItems[0]
	require.Equal(t, "baz", baz.Key())
	require.Equal(t, "baz value", baz.value)

	fooItems, has := l.indexed["foo"]
	require.True(t, has, "missing 'foo' items")
	require.Len(t, fooItems, 1, "there must be 1 'foo' item")

	foo := fooItems[0]
	require.Equal(t, "foo", foo.Key())
	require.Equal(t, "foo value", foo.value)

	require.NoError(t, err)
}
