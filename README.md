# Redis backed TODO app with Fn Golang functions 

## Switch to correct context

- `fn use context <your context name>`
- `fn ls apps`

## Common - create app

`fn create app --annotation oracle.com/oci/subnetIds='["ocid1.subnet.oc1.phx.aaaaaaaaghmsma7mpqhqdhbgnby25u2zo4wqlrrcskvu7jg56dryxt3hgvka"]' --config REDIS_HOST=your-redis-ip --config REDIS_PORT=your-redis-port --syslog-url=tcp://s3cr3t.papertrailapp.com:19407 todoapp`

> `--syslog-url` is optional. Use your own!

### To point to a different Redis instance

Jsut udpate the app configuration as below

`fn update app --annotation oracle.com/oci/subnetIds='["ocid1.subnet.oc1.phx.aaaaaaaaghmsma7mpqhqdhbgnby25u2zo4wqlrrcskvu7jg56dryxt3hgvka"]' --config REDIS_HOST=your-redis-ip --config REDIS_PORT=your-redis-port todoapp`

## Create TODO function

`cd create` and then deploy with `fn -v deploy --app todoapp`

> Go to OCIR and make sure your repo is converted to PUBLIC

### create multiple TODOs

- `echo -n 'do this' | DEBUG=1 fn invoke todoapp createtodo`

> Expected response - `{"Status":"SUCCESS","Message":"Created TODO with ID 1", "Todoid":"1"}`

- `echo -n 'do that' | DEBUG=1 fn invoke todoapp createtodo`

### check TODOs 

Use Redis in Docker (or install `redis-cli` locally) - `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 hgetall todo:1` (repeat same for other TODOs with Ids `2` and so on....)

You should see

    todoid
    2
    title
    do that
    completed
    false

## Get TODO function

`cd get` and deploy with `fn -v deploy --app todoapp`

> Go to OCIR and make sure your repo is converted to PUBLIC

### get all TODOs

`DEBUG=1 fn invoke todoapp gettodos`

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

`echo -n 'completed' | DEBUG=1 fn invoke todoapp gettodos`

You will get an empty `[]` array since all the TODOs are pending (see next)

Check in Redis `docker run --rm redis redis-cli -h 129.213.91.171 -p 6379 lrange todos:completed 0 10`

> Expect an empty reply

### get pending TODOs

`echo -n 'pending' | DEBUG=1 fn invoke todoapp gettodos`

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

`cd toggle` and deploy with `fn -v deploy --app todoapp`

> Go to OCIR and make sure your repo is converted to PUBLIC

### mark completed

`echo -n '{"todoid": "1", "completed":"true"}' | DEBUG=1 fn invoke todoapp toggletodo`

Expected response - `{"Status":"SUCCESS","Message":"Toggled TODO with ID 1 to true"}`

it should show up in completed list - `echo -n 'completed' | DEBUG=1 fn invoke todoapp gettodos`

	[
	  {
	    "Todoid": "1",
	    "Title": "do this",
	    "Completed": "true"
	  }
	]

it should NOT show up in pending list - `echo -n 'pending' | DEBUG=1 fn invoke todoapp gettodos`

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

`echo -n '{"todoid": "1", "completed":"false"}' | DEBUG=1 fn invoke todoapp toggletodo`

Expected response - `{"Status":"SUCCESS","Message":"Toggled TODO with ID 1 to false"}`

it should NOT show up in completed list - `echo -n 'completed' | DEBUG=1 fn invoke todoapp gettodos`

it should show up in pending list - `echo -n 'pending' | DEBUG=1 fn invoke todoapp gettodos`

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

`cd edit` and deploy with `fn -v deploy --app todoapp`

> Go to OCIR and make sure your repo is converted to PUBLIC

`echo -n '{"todoid": "2", "title":"new title"}' | DEBUG=1 fn invoke todoapp edittodo`

Expected response `{"Status":"SUCCESS","Message":"Updated title for TODO 2"}`

Confirm - `DEBUG=1 fn invoke todoapp gettodos`

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

`cd delete` and deploy with `fn -v deploy --app todoapp`

`echo -n '1' | DEBUG=1 fn invoke todoapp deletetodo`

Response `{"Status":"SUCCESS","Message":"Deleted TODO 1"}`

Confirm - `DEBUG=1 fn invoke todoapp gettodos`

JSON reponse with TODO `1` removed

	[
	  {
	    "Todoid": "2",
	    "Title": "new title",
	    "Completed": "false"
	  }
	]
