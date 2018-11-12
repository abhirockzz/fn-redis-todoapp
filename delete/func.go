package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"

	fdk "github.com/fnproject/fdk-go"
	"github.com/go-redis/redis"
)

func main() {

	fdk.Handle(fdk.HandlerFunc(deleteTODOHandler))

}

const todoHashNamePrefix string = "todo:"
const pendingTodosListName string = "todos:pending"
const completedTodosListName string = "todos:completed"

const failedStatus string = "FAILED"
const successStatus string = "SUCCESS"

var client *redis.Client

func deleteTODOHandler(ctx context.Context, in io.Reader, out io.Writer) {
	redisHost := fdk.GetContext(ctx).Config()["REDIS_HOST"]
	redisPort := fdk.GetContext(ctx).Config()["REDIS_PORT"]

	if client == nil {
		log.Println("Connecting to Redis...")
		opts := redis.Options{Addr: redisHost + ":" + redisPort}
		client = redis.NewClient(&opts)
		_, conErr := client.Ping().Result()

		if conErr != nil {
			json.NewEncoder(out).Encode(result{Status: failedStatus, Message: conErr.Error()})
			return
		}

		log.Println("Connected to Redis")
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	todoID := buf.String()

	if todoID == "" {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO ID cannot be empty!"})
		return
	}

	if client.Exists(todoHashNamePrefix+todoID).Val() == 0 {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO with ID " + todoID + " does not exist"})
		return
	}

	currentCompletedStatus := client.HGetAll(todoHashNamePrefix + todoID).Val()["completed"]
	log.Println("Current status for TODO " + todoID + " is " + currentCompletedStatus)

	txPipe := client.TxPipeline()

	//use pipeline to delete TODO HASH and remove from respective LIST (completed or pending)

	delResult := txPipe.Del(todoHashNamePrefix + todoID)

	var listName string
	if currentCompletedStatus == "true" {
		listName = completedTodosListName
	} else if currentCompletedStatus == "false" {
		listName = pendingTodosListName
	} else {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "Current completed status is invalid " + currentCompletedStatus})
		return
	}

	lremResult := txPipe.LRem(listName, 0, todoID)

	_, execErr := txPipe.Exec()

	if execErr != nil {
		log.Println("Tx error " + execErr.Error())
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: execErr.Error()})
		return
	}

	if delResult.Err() == nil && lremResult.Err() == nil {
		json.NewEncoder(out).Encode(result{Status: successStatus, Message: "Deleted TODO " + todoID})
	} else {
		log.Println("DEL error " + delResult.Err().Error())
		log.Println("LREM error " + lremResult.Err().Error())
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "Failed to delete TODO " + todoID})
	}

}

type result struct {
	Status  string //`json:"status"`
	Message string //`json:"message,omitempty"`
}
