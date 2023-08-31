package configstore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/fsnotify/fsnotify"
)

func fileTreeProvider(s *Store, dirname string) {
	fileTree(s, dirname, false)
}

func fileTreeRefreshProvider(s *Store, dirname string) {
	fileTree(s, dirname, true)
}

func fileTree(s *Store, dirname string, refresh bool) {
	if dirname == "" {
		return
	}

	providername := buildProviderName("filetree", refresh, dirname)

	items, err := loadItems(dirname)
	if err != nil {
		errorProvider(s, providername, err)
		return
	}

	inmem := inMemoryProvider(s, providername)
	inmem.mut.Lock()
	inmem.items = items
	inmem.mut.Unlock()

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

				// We don't care about chmods
				if event.Op&fsnotify.Chmod != 0 {
					continue
				}

				// Add new path if it's a directory
				if event.Op&fsnotify.Create != 0 {
					if err := watchDirectory(watcher, event.Name); err != nil {
						logError(err)
					}
				}

				// We can't stat a deleted path, then we always
				// remove old path even if it's not a directory
				if event.Op&fsnotify.Remove != 0 {
					_ = watcher.Remove(event.Name)
				}

				items, err := loadItems(dirname)
				if err != nil {
					logError(err)
				} else {
					inmem.mut.Lock()
					inmem.items = items
					inmem.mut.Unlock()
					s.NotifyWatchers()
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				logError(err)
			}
		}
	}()

	if err := watchDirectory(watcher, dirname); err != nil {
		errorProvider(s, providername, err)
	}
}

func loadItems(dirname string) ([]Item, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, f := range files {
		filename := filepath.Join(dirname, f.Name())
		subitems, err := walk(filename, f)
		if err != nil {
			return nil, err
		}
		items = append(items, subitems...)
	}
	return items, nil
}

func isDir(filename string, f os.FileInfo) bool {
	var isDirSymlink bool
	if f.Mode()&os.ModeSymlink != 0 {
		link, err := filepath.EvalSymlinks(filename)
		if err != nil {
			return false
		}
		fLink, err := os.Stat(link)
		if err != nil {
			return false
		}
		if fLink.IsDir() {
			isDirSymlink = true
		}
		return isDirSymlink
	}
	return f.IsDir()
}

func watchDirectory(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if isDir(path, f) {
			if err := watcher.Add(path); err != nil {
				return err
			}
		}

		return nil
	})
}

func walk(filename string, f os.FileInfo) ([]Item, error) {
	if isDir(filename, f) {
		return browseDir([]Item{}, filename, f.Name())
	}

	it, err := readItem(filename, f.Name())
	it.key = transformKey(f.Name())
	if err != nil {
		return nil, err
	}
	return []Item{it}, nil
}

func browseDir(items []Item, path, basename string) ([]Item, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return items, err
	}

	for _, f := range files {
		filename := filepath.Join(path, f.Name())
		if isDir(filename, f) {
			var subItems []Item
			subItems, err = browseDir(subItems, filename, filepath.Join(basename, f.Name()))
			if err != nil {
				return nil, err
			}
			items = append(items, subItems...)
			continue
		}

		it1, err := readItem(filename, basename)
		if err != nil {
			return items, err
		}
		items = append(items, it1)

		it2 := newItem(filepath.Join(basename, f.Name()), it1.value)
		items = append(items, it2)
	}

	return items, nil
}

func readItem(path, basename string) (Item, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Item{}, err
	}
	return newItem(basename, string(content)), nil
}

func newItem(name, content string) Item {
	priority := int64(5)
	first, _ := utf8.DecodeRuneInString(name)
	if unicode.IsUpper(first) {
		priority = 10
	}
	return NewItem(name, content, priority)
}
