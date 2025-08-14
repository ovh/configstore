package configstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileProviderYAML(t *testing.T) {
	var s = NewStore()
	s.File("tests/fixtures/fileprovider/test.yaml")
	l, err := s.GetItemList()
	require.NoError(t, err)

	assert.Equal(t, 2, l.Len())

	for _, i := range l.Items {
		t.Logf("items: %s - %s", i.key, i.value)
	}

	i, has := l.indexed["my-config-key-1"]
	require.True(t, has, "missing 'my-config-key-1' items")
	require.Len(t, i, 1, "there must be 1 'my-config-key-1' item")

	i, has = l.indexed["my-config-key-2"]
	require.True(t, has, "missing 'my-config-key-2' items")
	require.Len(t, i, 1, "there must be 1 'my-config-key-2' item")

}

func TestFileProviderJSON(t *testing.T) {
	var s = NewStore()
	s.File("tests/fixtures/fileprovider/test.json")
	l, err := s.GetItemList()
	require.NoError(t, err)

	assert.Equal(t, 2, l.Len())

	for _, i := range l.Items {
		t.Logf("items: %s - %s", i.key, i.value)
	}

	i, has := l.indexed["my-config-key-1"]
	require.True(t, has, "missing 'my-config-key-1' items")
	require.Len(t, i, 1, "there must be 1 'my-config-key-1' item")

	i, has = l.indexed["my-config-key-2"]
	require.True(t, has, "missing 'my-config-key-2' items")
	require.Len(t, i, 1, "there must be 1 'my-config-key-2' item")

}
