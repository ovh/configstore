package configstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTreeProvider(t *testing.T) {
	var s = NewStore()
	s.FileTree("tests/fixtures/filetreeprovider")
	l, err := s.GetItemList()
	require.NoError(t, err)

	assert.Equal(t, 8, l.Len())

	for _, i := range l.Items {
		t.Logf("items: %s - %s", i.key, i.value)
	}

	barItems, has := l.indexed["bar"]
	require.True(t, has, "missing 'bar' items")
	require.Len(t, barItems, 2, "there must be 2 'bar' items")

	bar := barItems[0]
	require.Equal(t, "bar", bar.Key())
	require.Equal(t, "biz value", bar.value)

	bar = barItems[1]
	require.Equal(t, "bar", bar.Key())
	require.Equal(t, "buz value", bar.value)

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

	barbazItems, has := l.indexed["bar/barbaz"]
	require.True(t, has, "missing 'bar/barbaz' items")
	require.Len(t, fooItems, 1, "there must be 1 'bar/barbaz' item")

	barbaz_fooValue := barbazItems[0]
	require.Equal(t, "bar/barbaz", barbaz_fooValue.Key())
	require.Equal(t, "baz_foo_value", barbaz_fooValue.value)

	barbaz_fooItems, has := l.indexed["bar/barbaz/foo"]
	require.True(t, has, "missing 'bar/barbaz/foo' items")
	require.Len(t, fooItems, 1, "there must be 1 'bar/barbaz/foo' item")

	barbaz_fooValue = barbaz_fooItems[0]
	require.Equal(t, "bar/barbaz/foo", barbaz_fooValue.Key())
	require.Equal(t, "baz_foo_value", barbaz_fooValue.value)

}

func ExampleStore_FileTree() {
	var s = NewStore()
	s.FileTree("tests/fixtures/filetreeprovider2")
	l, err := s.GetItemList()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("tests/fixtures/filetreeprovider2 contains:")
	for _, i := range l.Items {
		fmt.Printf("%s - %s\n", i.key, i.value)
	}

	s = NewStore()
	s.FileTree("tests/fixtures/filetreeprovider")
	l, err = s.GetItemList()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("tests/fixtures/filetreeprovider contains:")
	for _, i := range l.Items {
		fmt.Printf("%s - %s\n", i.key, i.value)
	}

	// Output:
	// tests/fixtures/filetreeprovider2 contains:
	// database/dev - buz value
	// database/dev/buz - buz value
	// database/dev - fiz value
	// database/dev/fiz - fiz value
	// database/prod - prod foo value
	// database/prod/foo - prod foo value
	// foo - foo value
	// tests/fixtures/filetreeprovider contains:
	// bar/barbaz - baz_foo_value
	// bar/barbaz/foo - baz_foo_value
	// bar - biz value
	// bar/biz - biz value
	// bar - buz value
	// bar/buz - buz value
	// baz - baz value
	// foo - foo value
}
