package configstore

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/ghodss/yaml"
)

// Logs functions can be overriden
var LogErrorFunc = log.Printf
var LogInfoFunc = log.Printf

/*
** DEFAULT PROVIDERS IMPLEMENTATION
 */

func logError(err error) {
	if LogErrorFunc != nil {
		LogErrorFunc("error: %v", err)
	}
}

func errorProvider(s *Store, name string, err error) {
	logError(err)
	s.RegisterProvider(name, newErrorProvider(err))
}

func newErrorProvider(err error) Provider {
	return func() (ItemList, error) {
		return ItemList{}, err
	}
}

func fileProvider(s *Store, filename string) {
	file(s, filename, false, nil)
}

func fileRefreshProvider(s *Store, filename string) {
	file(s, filename, true, nil)
}

func fileCustomProvider(s *Store, filename string, fn func([]byte) ([]Item, error)) {
	file(s, filename, false, fn)
}

func fileCustomRefreshProvider(s *Store, filename string, fn func([]byte) ([]Item, error)) {
	file(s, filename, true, fn)
}

func file(s *Store, filename string, refresh bool, fn func([]byte) ([]Item, error)) {

	if filename == "" {
		return
	}

	providername := buildProviderName("file", refresh, filename)

	vals, err := readFile(filename, fn)
	if err != nil {
		errorProvider(s, providername, err)
		return
	}
	inmem := inMemoryProvider(s, providername)
	if LogInfoFunc != nil {
		LogInfoFunc("configuration from file: %s", filename)
	}
	inmem.Add(vals...)

	if !refresh {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errorProvider(s, providername, err)
		return
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case <-s.ctx.Done():
				return

			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				if event.Op&fsnotify.Write != 0 {
					vals, err := readFile(filename, fn)
					if err != nil {
						logError(err)
					} else {
						inmem.mut.Lock()
						inmem.items = vals
						inmem.mut.Unlock()
						s.NotifyWatchers()
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				logError(err)
			}
		}
	}()

	if err := watcher.Add(filename); err != nil {
		errorProvider(s, providername, err)
	}
}

func fileListProvider(s *Store, dirname string) {
	fileList(s, dirname, false)
}

func fileListRefreshProvider(s *Store, dirname string) {
	fileList(s, dirname, true)
}

func fileList(s *Store, dirname string, refresh bool) {
	if dirname == "" {
		return
	}

	providername := buildProviderName("filelist", refresh, dirname)

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		errorProvider(s, providername, err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			linkedFile, err := os.Stat(filepath.Join(dirname, file.Name()))
			if err != nil {
				errorProvider(s, providername, err)
				return
			}
			if linkedFile.IsDir() {
				continue
			}
		}

		if refresh {
			fileRefreshProvider(s, filepath.Join(dirname, file.Name()))
		} else {
			fileProvider(s, filepath.Join(dirname, file.Name()))
		}
	}
}

func readFile(filename string, fn func([]byte) ([]Item, error)) ([]Item, error) {
	vals := []Item{}
	b, err := os.ReadFile(filename)
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

func inMemoryProvider(s *Store, name string) *InMemoryProvider {
	inmem := &InMemoryProvider{}
	s.RegisterProvider(name, inmem.Items)
	return inmem
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

func envProvider(s *Store, prefix string) {

	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	prefixName := strings.ToUpper(prefix)
	if prefixName == "" {
		prefixName = "all"
	}
	inmem := inMemoryProvider(s, fmt.Sprintf("env:%s", prefixName))

	prefix = transformKey(prefix)

	for _, e := range os.Environ() {
		ePair := strings.SplitN(e, "=", 2)
		if len(ePair) <= 1 {
			continue
		}
		eTr := transformKey(ePair[0])
		if strings.HasPrefix(eTr, prefix) {
			inmem.Add(NewItem(strings.TrimPrefix(eTr, prefix), ePair[1], 15))
		}
	}

	// once all items have been added, we need to notify watchers in case the goroutine watching for
	// providers change already scanned the Items
	s.NotifyWatchers()
}

func buildProviderName(name string, refresh bool, parameter string) string {
	if refresh {
		return fmt.Sprintf("%s+refresh:%s", name, parameter)
	}
	return fmt.Sprintf("%s:%s", name, parameter)
}
