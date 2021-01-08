package configstore

import (
	"fmt"
	"io/ioutil"
	"os"
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

	var items []Item
	for _, f := range files {
		filename := filepath.Join(dirname, f.Name())
		subitems, err := walk(filename, f)
		if err != nil {
			errorProvider(s, providername, err)
			return
		}
		items = append(items, subitems...)
	}

	inmem := inMemoryProvider(s, providername)
	for _, it := range items {
		inmem.Add(it)
	}
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
