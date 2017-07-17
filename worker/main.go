package main

import (
	"fmt"
	"image"
	"os"

	"github.com/tsauvajon/go-microservices-poc/dataAccess"
	"github.com/tsauvajon/go-microservices-poc/task"
)

var (
	masterLocation       string
	storageLocation      string
	keyValueStoreAddress string
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return
	}

	keyValueStoreAddress = os.Args[1]

	masterLocation, err := dataAccess.GetValue(keyValueStoreAddress, "masterAddress")

	if err != nil {
		fmt.Println(err)
		return
	}

	storageLocation, err := dataAccess.GetValue(keyValueStoreAddress, "storageAddress")

	if err != nil {
		fmt.Println(err)
		return
	}
}

func getNewTask(masterAddress string) (task.Task, error) {
	return task.Task{}, nil
}

func getImageFromStorage(storageAddress string, t task.Task) (image.Image, error) {
	return nil, nil
}

func doWorkOnImage(img image.Image) image.Image {
	return nil
}

func sendImageToStorage(storageAddress string, t task.Task, img image.Image) error {
	return nil
}

func registerFinishedTask(masterAddress string, t task.Task) error {
	return nil
}
