# Golang microservices POC

https://jacobmartins.com/2016/03/14/web-app-using-microservices-in-go-part-1-design/

# Getting started

Open 5 terminals

cd ./client, ./keyValueStore, ./fileStorage, ./master, ./taskStore and `go build` each of them

### Start each service

``` bash
# start keyValueStore (will be hosted on :3330)
./keyValueStore

# connect the fileStorage
./fileStorage :3332 :3330

# connect the taskStore
./taskStore :3331 :3330

# connect the master (hosted on :3333)
./master :3333 :3330

# connect the client (will be hosted on :3334)
./client :3330
```