package configstore

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

func fileTreeProvider(s *Store, dirname string) {
	if dirname == "" {
		return
	}

	providername := fmt.Sprintf("filetree:%s", dirname)

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		errorProvider(s, providername, err)
		return
	}

	items := []Item{}

	for _, f := range files {
		filename := filepath.Join(dirname, f.Name())

		if f.IsDir() {
			items, err = browseDir(items, filename, f.Name())
			if err != nil {
				errorProvider(s, providername, err)
				return
			}
		} else {
			it, err := readItem(filename, f.Name())
			it.key = transformKey(f.Name())
			if err != nil {
				errorProvider(s, providername, err)
				return
			}
			items = append(items, it)
		}
	}

	inmem := inMemoryProvider(s, providername)
	for _, it := range items {
		inmem.Add(it)
	}
}

func browseDir(items []Item, path, basename string) ([]Item, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return items, err
	}

	for _, f := range files {
		filename := filepath.Join(path, f.Name())
		if f.IsDir() {
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
	content, err := ioutil.ReadFile(path)
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
