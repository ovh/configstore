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
			return items, fmt.Errorf("subdir %s: encountered nested directory %s, max 1 level of nesting", basename, f.Name())
		}
		it, err := readItem(filename, f.Name())
		if err != nil {
			return items, err
		}
		it.key = transformKey(basename)
		items = append(items, it)
		it.key = transformKey(filepath.Join(basename, f.Name()))
		items = append(items, it)
	}

	return items, nil
}

func readItem(path, basename string) (Item, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Item{}, err
	}
	priority := int64(5)
	first, _ := utf8.DecodeRuneInString(basename)
	if unicode.IsUpper(first) {
		priority = 10
	}
	return Item{value: string(content), priority: priority}, nil
}
