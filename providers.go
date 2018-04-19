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

func ErrorProvider(name string, err error) {
	RegisterProvider(name, func() (ItemList, error) { return ItemList{}, err })
}

func File(filename string) {
	file(filename, false)
}

func FileRefresh(filename string) {
	file(filename, true)
}

func file(filename string, refresh bool) {

	if filename == "" {
		return
	}

	providername := fmt.Sprintf("file:%s", filename)

	last := time.Now()
	vals, err := readFile(filename)
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
				vals, err := readFile(filename)
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

func readFile(filename string) ([]Item, error) {
	vals := []Item{}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &vals)
	if err != nil {
		return nil, err
	}
	return vals, nil
}

type InMemoryProvider struct {
	items []Item
	mut   sync.Mutex
}

func (inmem *InMemoryProvider) Add(s ...Item) *InMemoryProvider {
	inmem.mut.Lock()
	defer inmem.mut.Unlock()
	inmem.items = append(inmem.items, s...)
	return inmem
}

func (inmem *InMemoryProvider) Items() (ItemList, error) {
	inmem.mut.Lock()
	defer inmem.mut.Unlock()
	return ItemList{Items: inmem.items}, nil
}

func InMemory(name string) *InMemoryProvider {
	inmem := &InMemoryProvider{}
	RegisterProvider(name, inmem.Items)
	return inmem
}
