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
	fdk.Handle(fdk.HandlerFunc(getTODOsHandler))

}

const todoIDCounterName string = "TodoIDs"
const todoHashNamePrefix string = "todo:"
const pendingTodosListName string = "todos:pending"
const completedTodosListName string = "todos:completed"

const failedStatus string = "FAILED"
const successStatus string = "SUCCESS"

var client *redis.Client

func getTODOsHandler(ctx context.Context, in io.Reader, out io.Writer) {
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
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	filter := buf.String()

	if filter == "" {
		getAllTODOs(in, out)
	} else if filter == "completed" || filter == "pending" {
		getTODOsFiltered(filter, in, out)
	} else {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: "Invalid filter " + filter})
		return
	}

}

func getAllTODOs(in io.Reader, out io.Writer) {

	log.Println("Getting all TODOs...")

	//we jsut scan once with a count of 1000 (assuming num todos is not more than this)
	//and accept the result as the number of TODOs
	keys, _, _ := client.Scan(0, todoHashNamePrefix+"*", 1000).Result()
	numTODOs := len(keys)
	log.Println("Total TODOs ", numTODOs)

	var todoInfoMaps []*redis.StringStringMapCmd
	txPipe := client.TxPipeline()

	//pipeline avoids multiple HGETALL calls
	for _, todoHashName := range keys {
		todoInfoMaps = append(todoInfoMaps, txPipe.HGetAll(todoHashName))
	}

	/*for i := 1; i <= int(numTODOs); i++ {
		todoInfoMaps = append(todoInfoMaps, txPipe.HGetAll(todoHashNamePrefix+strconv.Itoa(i)))
	}*/

	_, execErr := txPipe.Exec()

	if execErr != nil {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: execErr.Error()})
		return
	}

	var todos []todo
	for _, todoInfoMap := range todoInfoMaps {
		aTodo := mapToTODOStruct(todoInfoMap.Val())
		log.Println("Got TODO ", aTodo)
		todos = append(todos, aTodo)
	}

	json.NewEncoder(out).Encode(todos)
}

func getTODOsFiltered(filter string, in io.Reader, out io.Writer) {

	var todoListName string
	if filter == "completed" {
		todoListName = completedTodosListName
	} else {
		todoListName = pendingTodosListName
	}

	log.Println("searching for TODOs in ", todoListName)

	//how many TODOs ?
	numTODOs := client.LLen(todoListName).Val()
	todoIDs := client.LRange(todoListName, 0, numTODOs).Val()

	var todoInfoMaps []*redis.StringStringMapCmd
	txPipe := client.TxPipeline()

	//pipeline avoids 'numTODOs' number of calls
	for _, todoID := range todoIDs {
		todoInfoMaps = append(todoInfoMaps, txPipe.HGetAll("todo:"+todoID))
	}

	_, execErr := txPipe.Exec()

	if execErr != nil {
		json.NewEncoder(out).Encode(result{Status: failedStatus, Message: execErr.Error()})
		return
	}

	todos := make([]todo, 0) //todos won't be (returned as) null
	for _, todoInfoMap := range todoInfoMaps {
		aTodo := mapToTODOStruct(todoInfoMap.Val())
		log.Println("Got TODO ", aTodo)
		todos = append(todos, aTodo)
	}

	json.NewEncoder(out).Encode(todos)
}

func mapToTODOStruct(todoInfoMap map[string]string) todo {
	todo := todo{}
	todo.Todoid = todoInfoMap["todoid"]
	todo.Title = todoInfoMap["title"]
	todo.Completed = todoInfoMap["completed"]
	return todo
}

type todo struct {
	Todoid    string
	Title     string
	Completed string
}

type result struct {
	Status  string //`json:"status"`
	Message string //`json:"message,omitempty"`
}
