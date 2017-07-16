package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/tsauvajon/go-microservices-poc/registering"
)

const (
	// StateWorking : currently working on this image
	StateWorking = "working"
	// StateFinished : finished the work on this image
	StateFinished = "finished"
)

func main() {
	if !registering.RegisterInKeyValueStore("storageAddress") {
		return
	}

	http.HandleFunc("/sendImage", receiveImage)
	http.HandleFunc("/getImage", serveImage)
	http.ListenAndServe(":3332", nil)
}

func receiveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondOnlyXAccepted(w, "POST")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	id := values.Get("id")

	if len(id) == 0 {
		respondWithError(w, "invalid id")
		return
	}

	state := values.Get("state")

	if state != StateWorking && state != StateFinished {
		respondWithError(w, "invalid state")
	}

	_, err = strconv.Atoi(id)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	file, err := os.Create("/tmp/" + state + "/" + id + ".png")
	defer file.Close()

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(file, r.Body)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, "Success")
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondOnlyXAccepted(w, "GET")
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	id := values.Get("id")

	if len(id) == 0 {
		respondWithError(w, "invalid ID")
		return
	}

	state := values.Get("state")

	if state != StateWorking && state != StateFinished {
		respondWithError(w, "invalid state")
		return
	}

	// we check that the id is a number
	_, err = strconv.Atoi(id)

	if err != nil {
		respondWithError(w, "invalid ID")
		return
	}

	file, err := os.Open("tmp/" + state + "/" + id + ".png")
	defer file.Close()

	if err != nil {
		respondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(w, file)

	if err != nil {
		respondWithErrorStack(w, err)
		return
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
