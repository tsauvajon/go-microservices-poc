package main

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"sync"

	"time"

	"net/http"

	"encoding/json"
	"io/ioutil"

	"image/color"
	"image/png"

	"bytes"

	"errors"

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

	value, err := dataAccess.GetValue(keyValueStoreAddress, "masterAddress")

	masterLocation = value

	if err != nil {
		fmt.Println(err)
		return
	}

	value, err = dataAccess.GetValue(keyValueStoreAddress, "storageAddress")

	storageLocation = value

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
	response, err := http.Post("http://"+masterAddress+"/getNewTask", "text/plain", nil)

	if err != nil || response.StatusCode != http.StatusOK {
		fmt.Println("Error: ", "getNewTask => http.Post", err.Error(), response.Status)
		return task.Task{
			ID:    -1,
			State: -1,
		}, err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println("Error: ", "getNewTask => ioutil.ReadAll", err.Error())
		return task.Task{
			ID:    -1,
			State: -1,
		}, err
	}

	t := task.Task{}

	err = json.Unmarshal(data, &t)

	if err != nil {
		fmt.Println("Error: ", "getNewTask => json.Unmarshal", err.Error())
		return task.Task{
			ID:    -1,
			State: -1,
		}, err
	}

	return t, nil
}

func getImageFromStorage(storageAddress string, t task.Task) (image.Image, error) {
	response, err := http.Get("http://" + storageAddress + "/getImage?state=working&id=" + strconv.Itoa(t.ID))

	if err != nil || response.StatusCode != http.StatusOK {
		fmt.Println("Error: ", "getImageFromStorage => http.Get", err.Error(), response.Status)
		return nil, err
	}

	// img, err := png.Decode(response.Body)

	// if err != nil {
	// 	return nil, err
	// }

	// return img, nil

	return png.Decode(response.Body)
}

// invert reds and greens
func doWorkOnImage(img image.Image) image.Image {
	canvas := image.NewRGBA(img.Bounds())

	for i := 0; i < canvas.Rect.Max.X; i++ {
		for j := 0; j < canvas.Rect.Max.Y; j++ {
			r, g, b, a := img.At(i, j).RGBA()
			color := new(color.RGBA)
			color.R = uint8(g)
			color.G = uint8(r)
			color.B = uint8(b)
			color.A = uint8(a)
			canvas.Set(i, j, color)
		}
	}

	return canvas.SubImage(img.Bounds())
}

func sendImageToStorage(storageAddress string, t task.Task, img image.Image) error {
	data := []byte{}
	buffer := bytes.NewBuffer(data)

	err := png.Encode(buffer, img)

	if err != nil {
		fmt.Println("Error: ", "sendImageToStorage => png.Encode", err.Error())
		return err
	}

	id := strconv.Itoa(t.ID)

	response, err := http.Post("http://"+storageAddress+"/sendImage?state=finished&id="+id, "image/png", buffer)

	if err != nil {
		fmt.Println("Error: ", "sendImageToStorage => http.Post", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("Error: " + "unexpected response => " + response.Status)
	}

	return nil
}

func registerFinishedTask(masterAddress string, t task.Task) error {
	id := strconv.Itoa(t.ID)
	response, err := http.Post("http://"+masterAddress+"/registerFinishedTask?id="+id, "text/plain", nil)

	if err != nil {
		fmt.Println("Error: ", "registerFinishedTask => http.Post", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("Error: " + "unexpected response => " + response.Status)
	}

	return nil
}
