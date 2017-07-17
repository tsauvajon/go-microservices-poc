package dataAccess

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

/*
RegisterInKeyValueStore :
Register a service's address in the key-value Store
*/
func RegisterInKeyValueStore(key string) bool {
	if len(os.Args) < 3 {
		fmt.Println("Too few arguments")
		return false
	}

	// itself
	selfAddress := os.Args[1]
	keyValueStoreAddress := os.Args[2]

	// Todo : use body instead ...
	response, err := http.Post("http://"+keyValueStoreAddress+"/set?key="+key+"&value="+selfAddress, "", nil)

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
