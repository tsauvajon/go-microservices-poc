package main

import (
	"log"
	"net/http"
	"os"

	"fmt"

	"io/ioutil"

	"net/url"

	"encoding/json"

	"io"

	"github.com/tsauvajon/go-microservices-poc/dataAccess"
	"github.com/tsauvajon/go-microservices-poc/errorHandling"
	"github.com/tsauvajon/go-microservices-poc/task"
)

var (
	databaseLocation string
	storageLocation  string
	test             task.Task
)

func main() {
	if !dataAccess.RegisterInKeyValueStore("masterAddress") {
		return
	}

	keyValueStoreAddress := os.Args[2]

	value, err := dataAccess.GetValue(keyValueStoreAddress, "databaseAddress")

	databaseLocation = value

	if err != nil {
		fmt.Println(err)
		return
	}

	value, err = dataAccess.GetValue(keyValueStoreAddress, "storageAddress")

	storageLocation = value

	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/newImage", newImage)
	http.HandleFunc("/getImage", getImage)
	http.HandleFunc("/isReady", isReady)
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/registerTaskFinished", registerTaskFinished)

	http.ListenAndServe(":3333", nil)
}

func newImage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("newImage")

	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	response, err := http.Post("http://"+databaseLocation+"/newTask", "text/plain", nil)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	defer response.Body.Close()
	id, err := ioutil.ReadAll(response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Println("Image id :", id)

	_, err = http.Post("http://"+storageLocation+"/sendImage?id="+string(id)+"&state=working", "image", r.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, string(id))
}

func getImage(w http.ResponseWriter, r *http.Request) {
	log.Println("getImage")

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

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	requestedTask := task.Task{}

	json.Unmarshal(data, &requestedTask)

	if requestedTask.State == task.StatusFinished {
		fmt.Fprint(w, task.StatusFinished)
		return
	}

	fmt.Fprint(w, task.StatusInProgress)
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getNewTask")

	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	response, err := http.Post("http://"+databaseLocation+"/getNewTask", "text/plain", nil)

	if err != nil {
		fmt.Println("master :193", err.Error())
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(w, response.Body)

	if err != nil {
		fmt.Println("master :201", err.Error())
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Println("getNewTask => no error")
}

func registerTaskFinished(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
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

	response, err := http.Post("http://"+databaseLocation+"/finishTask?id="+id, "text/plain", nil)

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
