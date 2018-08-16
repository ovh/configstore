package configstore

import (
	"reflect"
	"testing"
	"time"

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
			NewItem("bool", "true", 0),
			NewItem("int", "-42", 0),
			NewItem("uint", "42", 0),
			NewItem("float", "42.42", 0),
			NewItem("duration", "42s", 0),
			NewItem("bytes", "Y29uZmlnc3RvcmU=", 0),
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
	ProviderLen := 13
	ProviderElementsKeys := []string{"sql", "other", "bool", "int", "uint", "float", "duration", "bytes"}

	items, err := GetItemList()
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)

	// Basics
	assert.Len(items.Items, ProviderLen)
	assert.ElementsMatch(items.Keys(), ProviderElementsKeys)

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
	assert.ElementsMatch(items.Keys(), ProviderElementsKeys)

	// Basic sanity test (no error on empty set)
	assert.Len(Filter().Slice("fiewrhfgoiwerhgiroew").Rekey(RekeyFunc).Reorder(ROLowPrio).Squash().Apply(items).Items, 0)

	// Check filter decription
	assert.Equal(Filter().Slice("sql").Unmarshal(func() interface{} { return &DBItem{} }).String(), `sql: {"id":"","type":""}`)

	// CheckBool
	assert.Equal(must(Filter().GetItemValueBool("bool")), true)

	// CheckBool with chained calls
	assert.Equal(must(Filter().MustGetItem("bool").ValueBool()), true)

	// CheckInt
	assert.Equal(must(Filter().GetItemValueInt("int")), int64(-42))

	// CheckUint
	assert.Equal(must(Filter().GetItemValueUint("uint")), uint64(42))

	// CheckFloat
	assert.Equal(must(Filter().GetItemValueFloat("float")), float64(42.42))

	// CheckDuration
	assert.Equal(must(Filter().GetItemValueDuration("duration")), 42*time.Second)

	// CheckBytes
	assert.Equal(must(Filter().MustGetItem("bytes").ValueBytes()), []byte{99,111,110,102,105,103,115,116,111,114,101})

	// Check item not found
	_, err = items.GetItem("notfound")
	assert.Equal(mustType(err, ErrItemNotFound("")), true)

	_, err = items.GetItem("duration")
	assert.Equal(mustType(err, ErrItemNotFound("")), false)

	// Check uninitialized item list
	tmp, items := items, nil
	_, err = items.GetItem("duration")
	assert.Equal(mustType(err, ErrUninitializedItemList("")), true)
	items = tmp

	_, err = items.GetItem("duration")
	assert.Equal(mustType(err, ErrUninitializedItemList("")), false)

	// Check ambigous item
	_, err = items.GetItem("sql")
	assert.Equal(mustType(err, ErrAmbiguousItem("")), true)

	_, err = items.GetItem("duration")
	assert.Equal(mustType(err, ErrAmbiguousItem("")), false)
}

func mustValue(i Item) string {
	v, err := i.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return i
}

func mustType(a interface{}, b interface{}) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}
