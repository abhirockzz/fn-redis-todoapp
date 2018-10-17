package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	fdk "github.com/fnproject/fdk-go"
	"github.com/go-redis/redis"
)

var redisHost string
var redisPort string

func main() {

	redisHost = os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort = os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	fdk.Handle(fdk.HandlerFunc(toggleTODOHandler))

}

const todoIDCounterName string = "TodoIDs"
const todoHashNamePrefix string = "todo:"
const completedTodosListName string = "todos:completed"
const pendingTodosListName string = "todos:pending"
const failedStatus string = "FAILED"
const successStatus string = "SUCCESS"

var client *redis.Client

func toggleTODOHandler(ctx context.Context, in io.Reader, out io.Writer) {

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

	var todoToggleInfo toggleInfo
	json.NewDecoder(in).Decode(&todoToggleInfo)
	log.Println("Toggle info ", todoToggleInfo)

	todoID := todoToggleInfo.Todoid
	completed := todoToggleInfo.Completed

	if todoID == "" {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO ID cannot be empty!"})
		return
	}

	if client.Exists(todoHashNamePrefix+todoID).Val() == 0 {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO with ID " + todoID + " does not exist"})
		return
	}

	//let's check whether the current completed status is the same as what's being passed
	//if yes, no point proceeding

	currentCompletedStatus := client.HGet(todoHashNamePrefix+todoID, "completed").Val()

	if currentCompletedStatus == completed {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO with ID " + todoID + " is already " + completed})
		return
	}

	txPipe := client.TxPipeline()

	//toggling has 3 steps... using pipeline
	hsetResult := txPipe.HSet(todoHashNamePrefix+todoID, "completed", completed)

	var lremResult, rpushResult *redis.IntCmd
	if completed == "true" {
		log.Println("will mark todo " + todoID + " as complete")

		//delete from pending to completed list
		lremResult = txPipe.LRem(pendingTodosListName, 0, todoID)
		rpushResult = txPipe.RPush(completedTodosListName, todoID)

	} else if completed == "false" {
		log.Println("will mark todo " + todoID + " as pending")

		lremResult = txPipe.LRem(completedTodosListName, 0, todoID)
		rpushResult = txPipe.RPush(pendingTodosListName, todoID)
	} else {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "Invalid value for 'completed' " + completed})
		return
	}

	_, execErr := txPipe.Exec()

	if execErr != nil {
		log.Println("Tx error ", execErr.Error())
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: execErr.Error()})
		return
	}

	if hsetResult.Err() == nil && lremResult.Err() == nil && rpushResult.Err() == nil {
		log.Println("Toggled TODO " + todoID + " to " + completed)
		json.NewEncoder(out).Encode(result{Status: successStatus, Message: "Toggled TODO with ID " + todoID + " to " + completed})
	} else {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "Failed to toggle TODO with ID " + todoID})
	}

}

type toggleInfo struct {
	Todoid    string
	Completed string
}

type result struct {
	Status  string //`json:"status"`
	Message string //`json:"message,omitempty"`
}
