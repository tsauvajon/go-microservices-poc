package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/tsauvajon/go-microservices-poc/errorHandling"
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

	http.ListenAndServe(":3330", nil)
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandling.RespondOnlyXAccepted(w, "GET")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	if len(values.Get("key")) == 0 {
		errorHandling.RespondWithError(w, "Wrong input key")
		return
	}

	// Mutex lock
	keyValueStoreMutex.RLock()
	value := keyValueStore[values.Get("key")]
	keyValueStoreMutex.RUnlock()

	fmt.Fprint(w, value)
}

func set(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	key := values.Get("key")
	value := values.Get("value")

	if len(key) == 0 {
		errorHandling.RespondWithError(w, "Wrong input key")
		return
	}

	if len(value) == 0 {
		errorHandling.RespondWithError(w "Wrong input value")
		return
	}

	keyValueStoreMutex.Lock()
	keyValueStore[key] = value
	keyValueStoreMutex.Unlock()

	fmt.Fprint(w, "Success")
}

func remove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		errorHandling.RespondOnlyXAccepted(w, "DELETE")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	key := values.Get("key")

	if len(key) == 0 {
		errorHandling.RespondWithError(w, "Wrong input key")
		return
	}

	keyValueStoreMutex.Lock()
	delete(keyValueStore, key)
	keyValueStoreMutex.Unlock()

	fmt.Fprint(w, "Success")
}

func list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandling.RespondOnlyXAccepted(w, "GET")
		return
	}

	keyValueStoreMutex.RLock()
	for key, value := range keyValueStore {
		fmt.Fprint(w, key, ": ", value)
	}
	keyValueStoreMutex.RUnlock()
}
