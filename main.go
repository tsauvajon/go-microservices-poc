package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

var (
	keyValueStore      map[string]string
	keyValueStoreMutex sync.RWMutex
)

func main() {
	keyValueStore = make(map[string]string)
	keyValueStoreMutex = sync.RWMutex{}

	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/remove", remove)
	http.HandleFunc("/list", list)

	http.ListenAndServe(":3333", nil)
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Only GET accepted")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", err)
		return
	}

	if len(values.Get("key")) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Wrong input key")
		return
	}

	// Mutex lock
	keyValueStoreMutex.RLock()
	value := keyValueStore[string(values.Get("key"))]
	keyValueStoreMutex.RUnlock()

	fmt.Fprint(w, value)
}

func set(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", err)
		return
	}

	key := values.Get("key")
	value := values.Get("value")

	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Wrong input key")
		return
	}

	if len(value) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Wrong input value")
		return
	}

	keyValueStoreMutex.Lock()
	keyValueStore[string(key)] = string(value)
	keyValueStoreMutex.Unlock()

	fmt.Fprint(w, "Success")
}

func remove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Only DELETE accepted")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", err)
		return
	}

	key := values.Get("key")

	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Wrong input key")
		return
	}

	keyValueStoreMutex.Lock()
	delete(keyValueStore, key)
	keyValueStoreMutex.Unlock()

	fmt.Fprint(w, "Success")
}

func list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", "Only GET accepted")
		return
	}

	keyValueStoreMutex.RLock()
	for key, value := range keyValueStore {
		fmt.Fprint(w, key, ": ", value)
	}
	keyValueStoreMutex.RUnlock()
}
