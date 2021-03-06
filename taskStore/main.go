package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/tsauvajon/go-microservices-poc/dataAccess"
	"github.com/tsauvajon/go-microservices-poc/errorHandling"
	"github.com/tsauvajon/go-microservices-poc/task"
)

var (
	datastore                  map[int]task.Task
	datastoreMutex             sync.RWMutex
	oldestNotFinishedTask      int // can int overflow, use something bigger in production
	oldestNotFinishedTaskMutex sync.RWMutex
)

func main() {
	if !dataAccess.RegisterInKeyValueStore("databaseAddress") {
		return
	}

	datastore = make(map[int]task.Task)
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

func getByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandling.RespondOnlyXAccepted(w, "GET")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	strid := values.Get("id")

	if len(strid) == 0 {
		errorHandling.RespondWithError(w, "Invalid ID")
		return
	}

	id, err := strconv.Atoi(strid)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	datastoreMutex.RLock()
	isInError := id >= len(datastore)
	datastoreMutex.RUnlock()

	if isInError {
		errorHandling.RespondWithError(w, "This ID does not exist")
		return
	}

	datastoreMutex.RLock()
	value := datastore[id]
	datastoreMutex.RUnlock()

	response, err := json.Marshal(value)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, string(response))
}

func newTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	datastoreMutex.RLock()
	taskToAdd := task.Task{
		ID:    len(datastore),
		State: task.StatusNotStarted,
	}
	datastore[taskToAdd.ID] = taskToAdd
	datastoreMutex.RUnlock()

	fmt.Fprint(w, taskToAdd.ID)
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getNewTask")

	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	isInError := false

	datastoreMutex.RLock()
	if len(datastore) == 0 {
		isInError = true
	}
	datastoreMutex.RUnlock()

	if isInError {
		errorHandling.RespondWithError(w, "no available task")
		return
	}

	taskToSend := task.Task{
		ID:    -1,
		State: task.StatusNotStarted,
	}

	oldestNotFinishedTaskMutex.Lock()
	datastoreMutex.Lock()

	fmt.Println("oldestNotFinishedTask:", oldestNotFinishedTask)

	for i := oldestNotFinishedTask; i < len(datastore); i++ {
		fmt.Println("checking tasks. ID:", datastore[i].ID, "State:", datastore[i].State)
		if i == oldestNotFinishedTask && datastore[i].State == task.StatusFinished {
			oldestNotFinishedTask++
			continue
		}

		if datastore[i].State == task.StatusNotStarted {
			taskToSend = datastore[i]
			break
		}
	}

	oldestNotFinishedTaskMutex.Unlock()
	datastoreMutex.Unlock()

	if taskToSend.ID == -1 {
		errorHandling.RespondWithError(w, "no available task")
		return
	}

	id := taskToSend.ID

	go func() {
		time.Sleep(time.Minute * 2)
		datastoreMutex.Lock()
		datastore[id] = task.Task{
			ID:    id,
			State: task.StatusNotStarted,
		}
		// set oldestNotFinishedTask to id ?
		datastoreMutex.Unlock()
	}()

	// taskToSend.ID: 0 taskToSend.State 0
	fmt.Println("taskToSend.ID:", taskToSend.ID, "taskToSend.State", taskToSend.State)

	response, err := json.Marshal(taskToSend)

	if err != nil {
		fmt.Println("taskStore :180 : ", err.Error())
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Println("getNewTask => ", string(response))

	fmt.Fprint(w, string(response))
}

func finishTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("finishTask")

	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		fmt.Println("taskStore :196 : ", err.Error())
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	strid := values.Get("id")

	if len(strid) == 0 {
		errorHandling.RespondWithError(w, "Invalid ID")
		return
	}

	id, err := strconv.Atoi(strid)

	if err != nil {
		fmt.Println("taskStore :211 : ", err.Error())
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	updatedTask := task.Task{
		ID:    id,
		State: task.StatusFinished,
	}

	fmt.Println("updating task => ID:", updatedTask.ID, "State:", updatedTask.State)

	isInError := false

	datastoreMutex.RLock()

	if datastore[id].State != task.StatusInProgress {
		isInError = true
	} else {
		datastore[id] = updatedTask
		fmt.Println("datastore length:", len(datastore))
	}
	datastoreMutex.RUnlock()

	if isInError {
		errorHandling.RespondWithError(w, "wrong input")
	}

	fmt.Fprint(w, "Success")
}

func setByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	taskToSet := task.Task{}

	err = json.Unmarshal([]byte(data), &taskToSet)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	isInError := false

	datastoreMutex.RLock()
	if taskToSet.ID >= len(datastore) || taskToSet.State < task.StatusNotStarted || taskToSet.State > task.StatusFinished {
		isInError = true
	} else {
		datastore[taskToSet.ID] = taskToSet
	}
	datastoreMutex.RUnlock()

	if isInError {
		errorHandling.RespondWithError(w, "wrong input")
	}

	fmt.Fprint(w, "Success")
}

func list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandling.RespondOnlyXAccepted(w, "GET")
		return
	}

	for key, value := range datastore {
		json, err := json.Marshal(value)

		if err != nil {
			errorHandling.RespondWithErrorStack(w, err)
			return
		}
		fmt.Fprintln(w, key, ": ", json)
	}
}
