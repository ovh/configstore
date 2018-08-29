package configstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

// ErrorProvider registers a configstore provider which always returns an error.
func ErrorProvider(name string, err error) {
	RegisterProvider(name, func() (ItemList, error) { return ItemList{}, err })
}

// File registers a configstore provider which reads from the file given in parameter (static content).
func File(filename string) {
	file(filename, false, nil)
}

// FileRefresh registers a configstore provider which readfs from the file given in parameter (provider watches file stat for auto refresh, watchers get notified).
func FileRefresh(filename string) {
	file(filename, true, nil)
}

// FileCustom registers a configstore provider which reads from the file given in parameter, and loads the content using the given unmarshal function
func FileCustom(filename string, fn func([]byte) ([]Item, error)) {
	file(filename, false, fn)
}

// FileCustomRefresh registers a configstore provider which reads from the file given in parameter, and loads the content using the given unmarshal function; and watches file stat for auto refresh
func FileCustomRefresh(filename string, fn func([]byte) ([]Item, error)) {
	file(filename, true, fn)
}

func file(filename string, refresh bool, fn func([]byte) ([]Item, error)) {

	if filename == "" {
		return
	}

	providername := fmt.Sprintf("file:%s", filename)

	last := time.Now()
	vals, err := readFile(filename, fn)
	if err != nil {
		ErrorProvider(providername, err)
		return
	}
	inmem := InMemory(providername)
	logrus.Infof("Configuration from file: %s", filename)
	inmem.Add(vals...)

	if refresh {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			for range ticker.C {
				finfo, err := os.Stat(filename)
				if err != nil {
					continue
				}
				if finfo.ModTime().After(last) {
					last = finfo.ModTime()
				} else {
					continue
				}
				vals, err := readFile(filename, fn)
				if err != nil {
					continue
				}
				inmem.mut.Lock()
				inmem.items = vals
				inmem.mut.Unlock()
				NotifyWatchers()
			}
		}()
	}
}

func readFile(filename string, fn func([]byte) ([]Item, error)) ([]Item, error) {
	vals := []Item{}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if fn != nil {
		return fn(b)
	}
	err = yaml.Unmarshal(b, &vals)
	if err != nil {
		return nil, err
	}
	return vals, nil
}

// InMemoryProvider implements an in-memory configstore provider.
type InMemoryProvider struct {
	items []Item
	mut   sync.Mutex
}

// Add appends an item to the in-memory list.
func (inmem *InMemoryProvider) Add(s ...Item) *InMemoryProvider {
	inmem.mut.Lock()
	defer inmem.mut.Unlock()
	inmem.items = append(inmem.items, s...)
	return inmem
}

// Items returns the in-memory item list. This is the function that gets called by configstore.
func (inmem *InMemoryProvider) Items() (ItemList, error) {
	inmem.mut.Lock()
	defer inmem.mut.Unlock()
	return ItemList{Items: inmem.items}, nil
}

// InMemory registers an InMemoryProvider with a given arbitrary name and returns it.
// You can append any number of items to it, see Add().
func InMemory(name string) *InMemoryProvider {
	inmem := &InMemoryProvider{}
	RegisterProvider(name, inmem.Items)
	return inmem
}
