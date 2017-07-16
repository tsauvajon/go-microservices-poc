package main

import (
	"net/http"
	"os"

	"errors"
	"fmt"

	"io/ioutil"

	"net/url"

	"encoding/json"

	"io"

	"github.com/tsauvajon/go-microservices-poc/errorHandling"
	"github.com/tsauvajon/go-microservices-poc/registering"
	"github.com/tsauvajon/go-microservices-poc/task"
)

var (
	databaseLocation string
	storageLocation  string
	test             task.Task
)

func main() {
	if !registering.RegisterInKeyValueStore("masterAddress") {
		return
	}

	keyValueStoreAddress := os.Args[2]

	value, err := getValueFromKeyValueStore(keyValueStoreAddress, "databaseAddress")

	if err != nil {
		fmt.Println(err)
	}

	databaseLocation = value

	value, err = getValueFromKeyValueStore(keyValueStoreAddress, "storageAddress")

	if err != nil {
		fmt.Println(err)
	}

	storageLocation = value

	fmt.Println(databaseLocation, storageLocation)

	http.HandleFunc("newImage", newImage)
	http.HandleFunc("getImage", getImage)
	http.HandleFunc("isReady", isReady)
	http.HandleFunc("getNewTask", getNewTask)
	http.HandleFunc("registerTaskFinished", registerTaskFinished)

	http.ListenAndServe(":3333", nil)
}

func newImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	response, err := http.Post("http://"+databaseLocation+"/newTask", "text/plain", nil)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	id, err := ioutil.ReadAll(response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	_, err = http.Post("http://"+storageLocation+"/sendImage?id="+string(id)+"&state=working", "image", r.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, string(id))
}

func getImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandling.RespondOnlyXAccepted(w, "GET")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	id := values.Get("id")

	if len(id) == 0 {
		errorHandling.RespondWithError(w, "invalid ID")
		return
	}

	response, err := http.Get("http://" + storageLocation + "/getImage?id=" + id + "&state=finished")

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(w, response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}
}

func isReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandling.RespondOnlyXAccepted(w, "GET")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	id := values.Get("id")

	if len(id) == 0 {
		errorHandling.RespondWithError(w, "invalid ID")
		return
	}

	response, err := http.Get("http://" + databaseLocation + "/getById?" + id)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	requestedTask := task.Task{}

	json.Unmarshal(data, &requestedTask)

	if requestedTask.State == task.StatusFinished {
		fmt.Fprint(w, "1")
		return
	}

	fmt.Fprint(w, "0")
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
}

func registerTaskFinished(w http.ResponseWriter, r *http.Request) {
}

func getValueFromKeyValueStore(address, key string) (string, error) {
	response, err := http.Get("http://" + address + "/get?key=" + key)

	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println(response.Body)
		return "", errors.New("Error: can't get the database address")
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	return string(data), nil
}
