package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"

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
	fdk.Handle(fdk.HandlerFunc(createTODOHandler))

}

const todoIDCounterName string = "TodoIDs"
const todoHashNamePrefix string = "todo:"
const pendingTodosListName string = "todos:pending"
const failedStatus string = "FAILED"
const successStatus string = "SUCCESS"

var client *redis.Client

func createTODOHandler(ctx context.Context, in io.Reader, out io.Writer) {

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
	todoTitle := buf.String()

	if todoTitle == "" {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO title cannot be empty!"})
		return
	}

	log.Println("TODO ", todoTitle)
	todoID, incrErr := client.Incr(todoIDCounterName).Result()
	if incrErr != nil {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: incrErr.Error()})
		return
	}
	log.Println("Generated TODO ID ", todoID)

	todoInfo := make(map[string]interface{})

	todoInfo["todoid"] = todoID
	todoInfo["title"] = todoTitle
	todoInfo["completed"] = "false"

	txPipe := client.TxPipeline()

	//we can execute below commands as a pipeline Tx to avoid 2 round trips
	hmsetResult := txPipe.HMSet(todoHashNamePrefix+strconv.Itoa(int(todoID)), todoInfo)
	lpushResult := txPipe.LPush(pendingTodosListName, strconv.Itoa(int(todoID)))

	_, execErr := txPipe.Exec()

	if execErr != nil {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: execErr.Error()})
		return
	}

	if hmsetResult.Err() == nil && lpushResult.Err() == nil {
		log.Println("created todo:" + strconv.Itoa(int(todoID)))
		log.Println("added to pending TODOs list...")

		json.NewEncoder(out).Encode(result{Status: successStatus, Message: "Created TODO with ID " + strconv.Itoa(int(todoID)), Todoid: strconv.Itoa(int(todoID))})
	} else {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "failed to create TODO with ID " + strconv.Itoa(int(todoID))})
	}
}

type result struct {
	Status  string //`json:"status"`
	Message string //`json:"message,omitempty"`
	Todoid  string `json:"Todoid,omitempty"`
}
