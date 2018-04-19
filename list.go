package configstore

import (
	"fmt"
	"sort"
)

// ItemList is a list of items which can be manipulated by an ItemFilter
type ItemList struct {
	Items   []Item
	indexed map[string][]Item
}

// GetItemList retrieves the full item list, merging the results from all providers.
// It does NOT cache, it's the responsability of the providers to keep an in-ram representation if desired.
func GetItemList() (*ItemList, error) {

	pMut.Lock()
	defer pMut.Unlock()

	ret := &ItemList{}

	for n, p := range providers {
		l, err := p()
		if err != nil {
			return nil, fmt.Errorf("configstore: provider '%s': %s", n, err)
		}
		ret.Items = append(ret.Items, l.Items...)
	}
	return ret.index(), nil
}

// GetItem retrieves the full item list, merging the results from all providers, then returns a single item by key.
func GetItem(key string) (Item, error) {
	items, err := GetItemList()
	if err != nil {
		return Item{}, err
	}
	return items.GetItem(key)
}

// Returns a list of the different keys present in the item list.
func (s *ItemList) Keys() []string {
	if s == nil {
		return nil
	}

	ret := []string{}
	for k := range s.indexed {
		ret = append(ret, k)
	}

	return ret
}

// Returns a single item, by key.
// If 0 or >=2 items are present with that key, it will return an error.
func (s *ItemList) GetItem(key string) (Item, error) {

	if s == nil {
		return Item{}, fmt.Errorf("configstore: get '%s': non-initialized item list", key)
	}

	l := (&ItemFilter{}).Slice(key).Apply(s)

	switch len(l.Items) {
	case 0:
		return Item{}, fmt.Errorf("configstore: get '%s': no item found", key)
	case 1:
		return l.Items[0], nil

	}
	return Item{}, fmt.Errorf("configstore: get '%s': ambiguous, %d items share that key", key, len(l.Items))
}

// Implements sort.Interface.
// NOT CONCURRENT SAFE.
func (s *ItemList) Len() int {
	return len(s.Items)
}

// Implements sort.Interface
// NOT CONCURRENT SAFE.
func (s *ItemList) Less(i, j int) bool {
	s1 := s.Items[i]
	s2 := s.Items[j]
	return s1.priority > s2.priority
}

// Implements sort.Interface
// NOT CONCURRENT SAFE.
func (s *ItemList) Swap(i, j int) {
	s1 := s.Items[i]
	s2 := s.Items[j]
	s.Items[i] = s2
	s.Items[j] = s1
}

// Indexes the items of the list by key for easy access.
func (s *ItemList) index() *ItemList {
	if s.indexed != nil {
		return s
	}
	sort.Sort(s)
	s.indexed = map[string][]Item{}
	for _, sec := range s.Items {
		s.indexed[sec.key] = append(s.indexed[sec.key], sec)
	}
	return s
}
