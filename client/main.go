package main

import (
	"fmt"
	"os"

	"net/http"

	"github.com/tsauvajon/go-microservices-poc/dataAccess"
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
}

func handleTask(w http.ResponseWriter, r *http.Request) {
}

func handleCheckForReadiness(w http.ResponseWriter, r *http.Request) {
}

func serveImage(w http.ResponseWriter, r *http.Request) {
}
