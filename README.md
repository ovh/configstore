# configstore

The configstore library aims to facilitate configuration discovery and management.
It mixes configuration items coming from various (abstracted) data sources, called *providers*.

## Providers

Providers represent an abstract data source. Their only role is to return a list of *items*.

Some built-in implementations are available (in-memory store, file reader), but the library exposes a way to register a provider *factory*, to extend it and bridge with any other existing system.

Example mixing several providers
```go
    // custom provider with hardcoded values
    func MyProviderFunc() (configstore.ItemList, error) {
        ret := configstore.ItemList{
            Items: []configstore.Item{
                // an item has 3 components: key, value, priority
                // they are defined by the provider, but can be modified later by the library user
                configstore.NewItem("key1", `value1-higher-prio`, 6),
                configstore.NewItem("key1", `value1-lower-prio`, 5),
                configstore.NewItem("key2", `value2`, 5),
            },
        }
        return ret, nil
    }

    func main() {

        configstore.RegisterProvider("myprovider", MyProviderFunc)

        configstore.File("/path/to/file.txt")

        // blends items from all 3 sources
        items, err := configstore.GetItemList()
        if err != nil {
            panic(err)
        }

        for _, i := range items.Items {
            val, err := i.Value()
            if err != nil {
                panic(err)
            }
            fmt.Println(i.Key(), val, i.Priority())
        }
    }
```

## Items

An *item* is composed of 3 fields:
* **Key**: The name of the item. Does not have to be unique. The provider is responsible for giving a sensible initial value.
* **Value**: The content of the item. This can be either manipulated as a plain scalar string, or as a marshaled (JSON or YAML) blob for complex objects.
* **Priority**: An abstract integer value to use when priorizing between items sharing the same key. The provider is responsible for giving a sensible initial value.

## Item retrieval

When calling *configstore.GetItemList()*, the caller gets an *ItemList*.

This object contains all the configuration items. To manipulate it, you can use a *ItemFilter* object, which provides convenient helper functions to select and reorder the items.

All objects are safe to use even when the item list is empty.

Example of use:
```go
    func main() {
        items, err := configstore.GetItemList()
        if err != nil {
            panic(err)
        }

        // we start by building a filter to manipulate our configuration items
        // we will apply it on our items list later
        filter := configstore.Filter()

        // get the databases
        filter = filter.Slice("database")

        // now we have a list of database objects, let's assume the payload resembles this:
        // {"name": "foo", "ip": "192.168.0.1", "port": 5432, "type": "RO"}
        // {"name": "foo", "ip": "192.168.0.1", "port": 5433, "type": "RW"}
        // {"name": "bar", "ip": "192.168.0.1", "port": 5434, "type": "RO"}
        //
        // the "database" initial key provides too little information to extract the data relating to a specific DB
        // we need to drill down into the value

        // we need to unmarshal the JSON representation of the whole sublist
        // we pass a factory function that instantiates objects of the correct concrete type
        filter = filter.Unmarshal(func() interface{} { return &Database{} })

        // now we want to actually index and lookup by database name, instead of the generic "database"
        // we apply a rekey function that does payload inspection
        filter = filter.Rekey(rekeyByName)

        // we have duplicate elements: database "foo" is present twice
        // we want to favor the RW instance if possible
        // we apply a reordering function that does payload inspection
        filter = filter.Reorder(prioritizeRW)

        // we only need 1 of each, we squash to only keep the single highest priority of each key
        filter = filter.Squash()

        // now we have only 2 items left:
        // {"name": "foo", "ip": "192.168.0.1", "port": 5433, "type": "RW"}
        // {"name": "bar", "ip": "192.168.0.1", "port": 5434, "type": "RO"}

        // we can finally apply it on our list
        items = filter.Apply(items)

        // the same thing, more concise:
        filter = configstore.Filter().Slice("database").Unmarshal(func() interface{} { return &Database{} }).Rekey(rekeyByName).Reorder(prioritizeRW).Squash()
        items, err = filter.GetItemList() // shortcut: applies the filter to the full list from configstore.GetItemList()
        if err != nil {
            panic(err)
        }
        // declaring your filter separately like this lets you define it globally and execute it later
        // that way, you can use its description (String()) to generate usage information.
    }

    type Database struct {
        Name string `json:"name"`
        IP   string `json:"ip"`
        Port int    `json:"port"`
        Type string `json:"type"`
    }

    func rekeyByName(s *configstore.Item) string {
        i, err := s.Unmarshaled()
        // we see here the error that was produced when we called *ItemList.Unmarshal(...)*
        // we ignore it for now, it will be handled when the *main()* retrieves the object.
        if err == nil {
            return i.(*Database).Name
        }
        return s.Key()
    }

    func prioritizeRW(s *configstore.Item) int64 {
        i, err := s.Unmarshaled()
        if err == nil {
            if i.(*Database).Type == "RW" {
                return s.Priority() + 1
            }
        }
        return s.Priority()
    }
    
```
