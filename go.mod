module github.com/ovh/configstore

go 1.23.0

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/stretchr/testify v1.11.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v0.7.2 // goccy/yaml is lower-case sensitive with yaml/json tags
	v0.7.1 // JSON unmarshaling bug
	v0.7.0 // unmarshaling bug
)
