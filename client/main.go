package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"net/http"

	"net/url"

	"github.com/tsauvajon/go-microservices-poc/dataAccess"
	"github.com/tsauvajon/go-microservices-poc/errorHandling"
	"github.com/tsauvajon/go-microservices-poc/task"
)

const htmlPage = "<html><head><title>Upload file</title></head><body><form enctype=\"multipart/form-data\" action=\"submitTask\" method=\"post\"> <input type=\"file\" name=\"uploadfile\" /> <input type=\"submit\" value=\"upload\" /> </form> </body> </html>"

var (
	keyValueStoreAddress string
	masterLocation       string
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: ", "too few arguments")
		return
	}

	keyValueStoreAddress = os.Args[1]

	masterLocation, err := dataAccess.GetValue(keyValueStoreAddress, "masterAddress")

	if err != nil {
		fmt.Println(err)
		return
	}

	if len(masterLocation) == 0 {
		fmt.Println("Error: ", "master address' length is 0")
		return
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/submitTask", handleTask)
	http.HandleFunc("/isReady", handleCheckForReadiness)
	http.HandleFunc("/getImage", serveImage)
	http.ListenAndServe(":3334", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, htmlPage)
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandling.RespondOnlyXAccepted(w, "POST")
		return
	}

	if err := r.ParseMultipartForm(10000000); err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	file, _, err := r.FormFile("uploadFile")

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	response, err := http.Post("http://"+masterLocation+"/new", "image", file)

	if err != nil || response.StatusCode != http.StatusOK {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	fmt.Fprint(w, string(data))
}

func handleCheckForReadiness(w http.ResponseWriter, r *http.Request) {
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

	response, err := http.Get("http://" + masterLocation + "/isReady?id=" + id + "&state=finished")

	if err != nil || response.StatusCode != http.StatusOK {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	// hashtag la flemme de faire ça correctement
	switch string(data) {
	case string(task.StatusInProgress):
		fmt.Fprint(w, "Your image is not ready yet")
	case string(task.StatusFinished):
		fmt.Fprint(w, "Your image is ready")
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error: ", "wrong progression status")
	}
}

func serveImage(w http.ResponseWriter, r *http.Request) {
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

	response, err := http.Get("http://" + masterLocation + "/get?id=" + id + "&state=finished")

	if err != nil || response.StatusCode != http.StatusOK {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}

	_, err = io.Copy(w, response.Body)

	if err != nil {
		errorHandling.RespondWithErrorStack(w, err)
		return
	}
}
