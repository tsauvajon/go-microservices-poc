package main

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"sync"

	"time"

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

	threadCount, err := strconv.Atoi(os.Args[2])

	if err != nil {
		fmt.Println("Error: ", "couldn't parse thread count")
		return
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(threadCount)

	for i := 0; i < threadCount; i++ {
		go func() {
			for {
				task, err := getNewTask(masterLocation)

				if err != nil {
					fmt.Println("Error: ", err)
					fmt.Println("2s timeout")
					time.Sleep(time.Second * 2)
					continue
				}

				img, err := getImageFromStorage(storageLocation, task)

				if err != nil {
					fmt.Println("Error: ", err)
					fmt.Println("2s timeout")
					time.Sleep(time.Second * 2)
					continue
				}

				img = doWorkOnImage(img)

				err = sendImageToStorage(storageLocation, task, img)

				if err != nil {
					fmt.Println("Error: ", err)
					fmt.Println("2s timeout")
					time.Sleep(time.Second * 2)
					continue
				}

				err = registerFinishedTask(masterLocation, task)

				if err != nil {
					fmt.Println("Error: ", err)
					fmt.Println("2s timeout")
					time.Sleep(time.Second * 2)
					continue
				}
			}
		}()
	}

	waitGroup.Wait()
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
