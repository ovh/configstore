package configstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ProviderTest() (ItemList, error) {
	ret := ItemList{
		Items: []Item{
			NewItem("sql", `{"id":"42","type":"RO"}`, 1),
			NewItem("sql", `{"id":"42","type":"RO"}`, 1),
			NewItem("sql", `{"id":"42","type":"RW"}`, 1),
			NewItem("sql", `{"id":"47","type":"RO"}`, 0),
			NewItem("other", `low`, -3),
			NewItem("other", `mid`, -2),
			NewItem("other", `higher`, -1),
		},
	}
	return ret, nil
}

type DBItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func RekeyFunc(s *Item) string {
	// Retrieve unmarshaled object
	i, err := s.Unmarshaled()
	if err == nil {
		// Cast it to the type returned by the factory we passed to ItemList.Unmarshal(func())
		// Use the DB identifier as the new key
		return i.(*DBItem).ID
	}
	return ""
}

func ROLowPrio(s *Item) int64 {
	// Retrieve unmarshaled object
	i, err := s.Unmarshaled()
	if err == nil {
		// Cast it to the type returned by the factory we passed to ItemList.Unmarshal(func())
		// Force RW values as higher priority
		if i.(*DBItem).Type == "RW" {
			return s.Priority() + 1
		}
	}
	return s.Priority()
}

func TestStore(t *testing.T) {
	RegisterProvider("test", ProviderTest)
	items, err := GetItemList()
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)

	// Basics
	assert.Len(items.Items, 7)
	assert.ElementsMatch(items.Keys(), []string{"sql", "other"})

	// Ensure basic order
	assert.Equal(mustValue(Filter().Slice("other").Apply(items).Items[0]), "higher")

	// Simple filter
	assert.Len(Filter().Slice("other").Apply(items).Items, 3)
	assert.Len(Filter().Slice("sql").Apply(items).Items, 4)

	// Rekey on sql subset
	dbItems := Filter().Slice("sql").Unmarshal(func() interface{} { return &DBItem{} }).Apply(items)
	assert.ElementsMatch(Filter().Rekey(RekeyFunc).Apply(dbItems).Keys(), []string{"42", "47"})
	assert.Len(Filter().Rekey(RekeyFunc).Slice("42").Apply(dbItems).Items, 3)
	assert.Len(Filter().Rekey(RekeyFunc).Slice("47").Apply(dbItems).Items, 1)

	// Reorder on sql subset (RW > RO)
	// Ensure order
	assert.Equal(mustValue(Filter().Rekey(RekeyFunc).Reorder(ROLowPrio).Apply(dbItems).Items[0]), `{"id":"42","type":"RW"}`)
	// Squash
	assert.Len(Filter().Rekey(RekeyFunc).Reorder(ROLowPrio).Squash().Apply(dbItems).Items, 2)
	assert.Len(Filter().Rekey(RekeyFunc).Reorder(ROLowPrio).Squash().Slice("42").Apply(dbItems).Items, 1)

	// Basic squash without reorder
	assert.Len(Filter().Slice("other").Squash().Apply(items).Items, 1)
	assert.Equal(mustValue(Filter().Slice("other").Squash().Apply(items).Items[0]), "higher")

	// Ensure initial collection did not get affected
	assert.ElementsMatch(items.Keys(), []string{"sql", "other"})

	// Basic sanity test (no error on empty set)
	assert.Len(Filter().Slice("fiewrhfgoiwerhgiroew").Rekey(RekeyFunc).Reorder(ROLowPrio).Squash().Apply(items).Items, 0)

	// Check filter decription
	assert.Equal(Filter().Slice("sql").Unmarshal(func() interface{} { return &DBItem{} }).String(), `sql: {"id":"","type":""}`)
}

func mustValue(i Item) string {
	v, err := i.Value()
	if err != nil {
		panic(err)
	}
	return v
}
