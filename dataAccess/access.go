package dataAccess

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetValue : get the value associated with a key
func GetValue(address, key string) (string, error) {
	fmt.Println("Getting ", key, " from ", address)

	response, err := http.Get("http://" + address + "/get?key=" + key)

	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println(response.Body)
		return "", errors.New("Error: can't get the database address")
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	fmt.Println("Returning value: ", string(data))

	return string(data), nil
}
