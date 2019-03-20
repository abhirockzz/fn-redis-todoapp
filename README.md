# TODO app with Golang functions and Redis 

## Switch to correct context

- `fn use context <your context name>`
- `fn ls apps`

## Create an application

`fn create app --annotation oracle.com/oci/subnetIds='[<SUBNET_OCID(s)>]' --config REDIS_HOST=your-redis-ip --config REDIS_PORT=your-redis-port todoapp`

### To point to a different Redis instance...

... execute below commands to update the app configuration

`fn config app todoapp REDIS_HOST your-redis-ip` and `fn config app todoapp REDIS_PORT your-redis-port`

## To deploy your functions...

Clone this repository and change into the directory - `cd fn-redis-todoapp`

To deploy all the functions in one go - `fn -v deploy --app todoapp --all`

You can choose to deploy one function at a time. You can do so by changing into the function directory and invoking the `fn deploy` command. For e.g. to deploy the create todo function - `cd create` and then deploy with `fn -v deploy --app todoapp`

Once the functions have been deployed, you can test out the CRUD capabilities

## Create TODOs

- `echo -n 'do this' | fn invoke todoapp createtodo`

> Expected response - `{"Status":"SUCCESS","Message":"Created TODO with ID 1", "Todoid":"1"}`

- `echo -n 'do that' | fn invoke todoapp createtodo`

### check TODOs 

Use Redis in Docker (or install `redis-cli` locally) - `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 hgetall todo:1` (repeat same for other TODOs with Ids `2` and so on....)

You should see

    todoid
    2
    title
    do that
    completed
    false

## Get TODOs

### get all TODOs

`fn invoke todoapp gettodos`

You should see JSON output

	[
	  {
	    "Todoid": "2",
	    "Title": "do that",
	    "Completed": "false"
	  },
	  {
	    "Todoid": "1",
	    "Title": "do this",
	    "Completed": "false"
	  }
	]

### get completed TODOs

`echo -n 'completed' | fn invoke todoapp gettodos`

You will get an empty `[]` array since all the TODOs are pending (see next)

Check in Redis `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 lrange todos:completed 0 10`

> Expect an empty reply

### get pending TODOs

`echo -n 'pending' | fn invoke todoapp gettodos`

You should see JSON output

	[
	  {
	    "Todoid": "2",
	    "Title": "do that",
	    "Completed": "false"
	  },
	  {
	    "Todoid": "1",
	    "Title": "do this",
	    "Completed": "false"
	  }
	]

Check in Redis `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 lrange todos:pending 0 10`

    1
    2

## Toggle TODO (complete or pending)

### mark completed

`echo -n '{"todoid": "1", "completed":"true"}' | fn invoke todoapp toggletodo`

Expected response - `{"Status":"SUCCESS","Message":"Toggled TODO with ID 1 to true"}`

it should show up in completed list - `echo -n 'completed' | fn invoke todoapp gettodos`

	[
	  {
	    "Todoid": "1",
	    "Title": "do this",
	    "Completed": "true"
	  }
	]

it should NOT show up in pending list - `echo -n 'pending' | fn invoke todoapp gettodos`

	[
	  {
	    "Todoid": "2",
	    "Title": "do that",
	    "Completed": "false"
	  }
	]

Check in Redis `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 lrange todos:completed 0 10`

    1

Check in Redis `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 lrange todos:pending 0 10`

    2

Responses will be in-line with above results

### mark pending

`echo -n '{"todoid": "1", "completed":"false"}' | fn invoke todoapp toggletodo`

Expected response - `{"Status":"SUCCESS","Message":"Toggled TODO with ID 1 to false"}`

it should NOT show up in completed list - `echo -n 'completed' | fn invoke todoapp gettodos`

it should show up in pending list - `echo -n 'pending' | fn invoke todoapp gettodos`

	[
	  {
	    "Todoid": "2",
	    "Title": "do that",
	    "Completed": "false"
	  },
	  {
	    "Todoid": "1",
	    "Title": "do this",
	    "Completed": "false"
	  }
	]

## Edit TODO function (update title)

`echo -n '{"todoid": "2", "title":"new title"}' | fn invoke todoapp edittodo`

Expected response `{"Status":"SUCCESS","Message":"Updated title for TODO 2"}`

Confirm - `fn invoke todoapp gettodos`

JSON response with updated title for TODO

	[
	  {
	    "Todoid": "2",
	    "Title": "new title",
	    "Completed": "false"
	  },
	  {
	    "Todoid": "1",
	    "Title": "do this",
	    "Completed": "false"
	  }
	]

## Delete TODO function

`echo -n '1' | fn invoke todoapp deletetodo`

Response `{"Status":"SUCCESS","Message":"Deleted TODO 1"}`

Confirm - `fn invoke todoapp gettodos`

JSON reponse with TODO `1` removed

	[
	  {
	    "Todoid": "2",
	    "Title": "new title",
	    "Completed": "false"
	  }
	]
