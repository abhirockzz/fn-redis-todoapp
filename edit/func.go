package main

import (
	"context"
	"encoding/json"
	"io"
	"log"

	fdk "github.com/fnproject/fdk-go"
	"github.com/go-redis/redis"
)

//var redisHost string
//var redisPort string

func main() {

	/*redisHost = os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort = os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}*/
	fdk.Handle(fdk.HandlerFunc(editTODOHandler))

}

const todoHashNamePrefix string = "todo:"
const failedStatus string = "FAILED"
const successStatus string = "SUCCESS"

var client *redis.Client

func editTODOHandler(ctx context.Context, in io.Reader, out io.Writer) {
	redisHost := fdk.Context(ctx).Config["REDIS_HOST"]
	redisPort := fdk.Context(ctx).Config["REDIS_PORT"]

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

	var todoEditInfo editInfo
	json.NewDecoder(in).Decode(&todoEditInfo)
	log.Println("TODO edit info ", todoEditInfo)

	todoID := todoEditInfo.Todoid
	title := todoEditInfo.Title

	if todoID == "" || title == "" {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "Please supply correct value for TODO Id and/or title"})
		return
	}

	if client.Exists(todoHashNamePrefix+todoID).Val() == 0 {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "TODO with ID " + todoID + " does not exist"})
		return
	}

	hsetErr := client.HSet(todoHashNamePrefix+todoID, "title", title).Err()

	if hsetErr == nil {
		json.NewEncoder(out).Encode(result{Status: successStatus, Message: "Updated title for TODO " + todoID})
	} else {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: hsetErr.Error()})

	}
}

type editInfo struct {
	Todoid string
	Title  string
}

type result struct {
	Status  string //`json:"status"`
	Message string //`json:"message,omitempty"`
}
