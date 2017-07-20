package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/tsauvajon/go-microservices-poc/dataAccess"
	"github.com/tsauvajon/go-microservices-poc/errorHandling"
)

const (
	// StateWorking : currently working on this image
	StateWorking = "working"
	// StateFinished : finished the work on this image
	StateFinished = "finished"
)

func main() {
	if !dataAccess.RegisterInKeyValueStore("storageAddress") {
		return
	}

	http.HandleFunc("/sendImage", receiveImage)
	http.HandleFunc("/getImage", serveImage)
	http.ListenAndServe(":3332", nil)
}

func receiveImage(w http.ResponseWriter, r *http.Request) {
	log.Println("receiveImage")

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
		errorHandling.RespondWithError(w, "invalid id")
		return
	}

	state := values.Get("state")

	if state != StateWorking && state != StateFinished {
		errorHandling.RespondWithError(w, "invalid state")
		return
	}

	_, err = strconv.Atoi(id)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	file, err := os.Create("/tmp/" + state + "/" + id + ".png")
	defer file.Close()

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(file, r.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, "Success")
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	log.Println("serveImage")

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

	state := values.Get("state")

	if state != StateWorking && state != StateFinished {
		errorHandling.RespondWithError(w, "invalid state")
		return
	}

	// we check that the id is a number
	_, err = strconv.Atoi(id)

	if err != nil {
		errorHandling.RespondWithError(w, "invalid ID")
		return
	}

	file, err := os.Open("tmp/" + state + "/" + id + ".png")
	defer file.Close()

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(w, file)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}
}
