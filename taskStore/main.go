package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	// StatusNotStarted : this task isn't started yet
	StatusNotStarted = 0
	// StatusInProgress : this task is in progress
	StatusInProgress = 1
	// StatusFinished : this task is done
	StatusFinished = 2
)

/*
Task :
Consecutive IDs
state :
	0 – not started
	1 – in progress
	2 – finished
*/
type Task struct {
	ID    int `json:"id"`
	State int `json:"state"`
}

var (
	datastore                  map[int]Task
	datastoreMutex             sync.RWMutex
	oldestNotFinishedTask      int // can int overflow, use something bigger in production
	oldestNotFinishedTaskMutex sync.RWMutex
)

func main() {

	datastore = make(map[int]Task)
	datastoreMutex = sync.RWMutex{}

	oldestNotFinishedTask = 0
	oldestNotFinishedTaskMutex = sync.RWMutex{}

	http.HandleFunc("/getByID", getByID)
	http.HandleFunc("/newTask", newTask)
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/finishTask", finishTask)
	http.HandleFunc("/setByID", setByID)
	http.HandleFunc("/list", list)

	http.ListenAndServe(":3331", nil)
}

func registerInKeyValueStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Too few arguments")
		return false
	}

	// itself
	databaseAddress := os.Args[1]
	keyValueStoreAddress := os.Args[2]

	// Todo : use body instead ...
	response, err := http.Post("http://"+keyValueStoreAddress+"/set?key=databaseAddress&value="+databaseAddress, "", nil)

	if err != nil {
		fmt.Println(err)
		return false
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: ", "failure contacting the key-value store", string(data))
		return false
	}

	return true
}

func getByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondOnlyXAccepted(w, "GET")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	strid := values.Get("id")

	if len(strid) == 0 {
		respondWithError(w, "Invalid ID")
		return
	}

	id, err := strconv.Atoi(strid)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	datastoreMutex.RLock()
	isInError := id >= len(datastore)
	datastoreMutex.RUnlock()

	if isInError {
		respondWithError(w, "This ID does not exist")
		return
	}

	datastoreMutex.RLock()
	value := datastore[id]
	datastoreMutex.RUnlock()

	response, err := json.Marshal(value)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, string(response))
}

func newTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondOnlyXAccepted(w, "POST")
		return
	}

	datastoreMutex.RLock()
	taskToAdd := Task{
		ID:    len(datastore),
		State: StatusNotStarted,
	}
	datastore[taskToAdd.ID] = taskToAdd
	datastoreMutex.RUnlock()

	fmt.Fprint(w, taskToAdd.ID)
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondOnlyXAccepted(w, "GET")
		return
	}

	isInError := false

	datastoreMutex.RLock()
	if len(datastore) == 0 {
		isInError = true
	}
	datastoreMutex.RUnlock()

	if isInError {
		respondWithError(w, "no available task")
		return
	}

	taskToSend := Task{
		ID:    -1,
		State: StatusNotStarted,
	}

	oldestNotFinishedTaskMutex.Lock()
	datastoreMutex.Lock()

	for i := oldestNotFinishedTask; i < len(datastore); i++ {
		if i == oldestNotFinishedTask && datastore[i].State == StatusFinished {
			oldestNotFinishedTask++
			continue
		}

		if datastore[i].State == StatusNotStarted {
			taskToSend = datastore[i]
			break
		}
	}

	oldestNotFinishedTaskMutex.Unlock()
	datastoreMutex.Unlock()

	if taskToSend.ID == -1 {
		respondWithError(w, "no available task")
		return
	}

	id := taskToSend.ID

	go func() {
		time.Sleep(time.Minute * 2)
		datastoreMutex.Lock()
		datastore[id] = Task{
			ID:    id,
			State: StatusNotStarted,
		}
		// set oldestNotFinishedTask to id ?
		datastoreMutex.Unlock()
	}()

	response, err := json.Marshal(taskToSend)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, response)
}

func finishTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondOnlyXAccepted(w, "POST")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	strid := values.Get("id")

	if len(strid) == 0 {
		respondWithError(w, "Invalid ID")
		return
	}

	id, err := strconv.Atoi(strid)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	updatedTask := Task{
		ID:    id,
		State: StatusFinished,
	}

	isInError := false

	datastoreMutex.RLock()

	if datastore[id].State != StatusInProgress {
		isInError = true
	} else {
		datastore[id] = updatedTask
	}
	datastoreMutex.RUnlock()

	if isInError {
		respondWithError(w, "wrong input")
	}

	fmt.Fprint(w, "Success")
}

func setByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondOnlyXAccepted(w, "POST")
		return
	}

	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	taskToSet := Task{}

	err = json.Unmarshal([]byte(data), &taskToSet)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	isInError := false

	datastoreMutex.RLock()
	if taskToSet.ID >= len(datastore) || taskToSet.State < StatusNotStarted || taskToSet.State > StatusFinished {
		isInError = true
	} else {
		datastore[taskToSet.ID] = taskToSet
	}
	datastoreMutex.RUnlock()

	if isInError {
		respondWithError(w, "wrong input")
	}

	fmt.Fprint(w, "Success")
}

func list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondOnlyXAccepted(w, "GET")
		return
	}

	for key, value := range datastore {
		json, err := json.Marshal(value)

		if err != nil {
			respondWithErrorStack(w, err)
			return
		}
		fmt.Fprintln(w, key, ": ", json)
	}
}

func respondWithErrorStack(w http.ResponseWriter, err error) {
	respondWithError(w, err.Error())
}

func respondOnlyXAccepted(w http.ResponseWriter, x string) {
	respondWithError(w, "only "+x+" accepted")
}

func respondWithError(w http.ResponseWriter, reason string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error : ", reason)
}
