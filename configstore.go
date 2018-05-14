package configstore

import (
	"sync"
)

var (
	providers = map[string]Provider{}
	pMut      sync.Mutex
)

// A Provider retrieves config items and makes them available to the configstore,
// Their implementations can vary wildly (HTTP API, file, env, hardcoded test, ...)
// and their results will get merged by the configstore library.
// It's the responsability of the application using configstore to register suitable providers.
type Provider func() (ItemList, error)

// RegisterProvider registers a provider
func RegisterProvider(name string, f Provider) {
	pMut.Lock()
	defer pMut.Unlock()
	providers[name] = f
}

var (
	watchers    []chan struct{}
	watchersMut sync.Mutex
)

// Watch returns a channel which you can range over.
// You will get unblocked every time a provider notifies of a configuration change.
func Watch() chan struct{} {
	// buffer size == 1, notifications will never use a blocking write
	newCh := make(chan struct{}, 1)
	watchersMut.Lock()
	watchers = append(watchers, newCh)
	watchersMut.Unlock()
	return newCh
}

// NotifyWatchers is used by providers to notify of configuration changes.
// It unblocks all the watchers which are ranging over a watch channel.
func NotifyWatchers() {
	watchersMut.Lock()
	for _, ch := range watchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	watchersMut.Unlock()
}
