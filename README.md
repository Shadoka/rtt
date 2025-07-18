# rtt - rabbit test tool

## Why and what

`rtt` is a test tool for rabbit queues aimed to be used similarly to tools like Postman and Bruno for HTTP calls.
Contrary to those tools `rtt` is purely a CLI tool.
Also, `rtt` can create topologies that are predefined in `setup.json` files.

Developers can create JSON files containing queue data (exchanges, routing keys, queue names) and message data.
For an example of a simple file [see this example file](examples/testB/testB.json). Optionally, you can define [a schema location](examples/schemas/example-schema.schema.json) to validate the message data contained in the file.

Additionally, if the consumer responds with a reply message it is possible to define the expected response message, [as you can see in this example](examples/testA/testA.json).

The general connection information (credentials & rabbit connection data) are defined in a file called [setup.json](examples/setup.json). This file can also define the whole topology of the rabbit infrastructure and should be referenced from each rtt test file. This has the advantage of not duplicating the credentials for each test and also to create the rabbit topology independent of the execution of a test file.

## Available commands

| Command  | Description |
| ------------- | ------------- |
| [run](#running-tests)  | Run one or more test files  |
| [validate](#validating-message-data)  | Validate message payloads against a schema  |
| [setup](#creating-the-topology-or-just-testing-the-connection)  | Creates the RabbitMQ topology defined in a setup.json  |
| [purge](#purging-messages)  | Purge messages from one or more queues  |
| [peek](#peek-into-a-queue-non-consuming)  | Print out messages without consuming them  |
| [send](#send-a-message)  | Sends a message to a queue  |
| [namespace create](#creating-a-namespace)  | Create a namespace  |
| [namespace update](#updating-a-namespace)  | Update a namespace  |
| [namespace list](#list-available-namespaces)  | List available namespaces  |
| [namespace set](#set-default-namespace)  | Set default namespace  |
| [namespace delete](#delete-a-namespace)  | Delete a namespace  |

## Usage

If you are on windows you have to replace 'rtt' with 'rtt.exe'.

### Running tests

Executing a single test file.

```sh
rtt run examples/testA/testA.json
```

You can also execute a directory of test files.

```sh
rtt run examples/
```

`rtt` waits 5 seconds by default for expected response messages. You can tweak that timeout duration by -t (duration in seconds)

```sh
rtt run examples/ -t 2
```

### Validating message data

If you receive new schema definitions you might want to check if your old message data is still valid first.
You can do this for a single file.

```sh
rtt validate examples/testA/testA.json
```

You can also do this for a directory of files.

```sh
rtt validate examples/
```

An example of an error message:

```sh
test a: FAILURE
jsonschema validation failed with 'file:///C:/workspaces/rtt/examples/schemas/example-schema.schema.json#'
- at '': missing property 'secondNumber'
test b: SUCCESS
test c: SUCCESS
```

In this case I renamed a property to 'thirdNumber' and the schema validation rightfully notes that the required field 'secondNumber' is nowhere to be found.

### Creating the topology (or just testing the connection)

To explicitly create a rabbit topology (or just test the connection to rabbit) you can just use following command.

```sh
rtt setup examples/setup.json
```

This command is idempotent, so you can run this as many times as you want without breaking things.
> [!NOTE]
> This setup is done automatically when you execute a file or directory with `rtt run`. So there is no need to manually run it beforehand.

### Purging messages

`rtt` can also purge messages from either a specific queue (defined in a `setup.json`) or from all queues that are defined in a `setup.json`.

To purge messages from one queue you have to specify the queue name with either `-q` or `--queue`:

```sh
rtt purge examples/setup.json -q addition-no-reply
```

To purge all messages from all queues just execute the purge command:

```sh
rtt purge examples/setup.json
```

### Peek into a queue (non-consuming)

Use the `peek` command if you want to see some of the first messages queued in a queue. Those messages will not be consumed by this action.
`rtt` does not interpret the data and (currently) does not validate the payload against a possibly configured schema.

By default `peek` will only print the first message in a queue and connects to the default namespace. You can change both those settings via
parameters.

Default (where `addition` is the name of the queue):
```sh
rtt peek addition
```

With parameters:
```sh
rtt peek addition -n 3 -s examples/setup.json
```

### Send a message

With the `send` command you can send a single message into a specified queue. You can use the default namespace or specify a setup.json file.

```sh
rtt send addition '{"someKey":"someValue"}'
```

The -s parameter is supported to supply a specific namespace:

```sh
rtt send addition '{"someKey":"someValue"}' -s examples/setup.json
```

## Namespaces

`rtt` has namespaces that connects to specific environments defined by the user via `setup.json`. Those namespaces are referenced by an alias.
You can also set a default namespace which has the advantage that you can omit the `setup.json` parameter to some calls.
The namespace management is done via subcommands of a namespace command.

### Creating a namespace

To create an namespace you can call the create subcommand with an alias and a path to a setup.json.

```sh
rtt namespace create local ./examples/setup.json
```

### Updating a namespace

To update the connection information or topology of a rabbit environment you can call the update subcommand with the alias and the new setup.json.

```sh
rtt namespace update local ./examples/setup.json
```

### List available namespaces

To list all available namespace you can simply call the list subcommand.

```sh
rtt namespace list
```

### Set default namespace

Call the set subcommand with an existing alias to set that namespace as a default namespace.

```sh
rtt namespace set local
```

### Delete a namespace

To delete a namespace you can call the delete subcommand with the alias of the namespace to delete.

```sh
rtt namespace delete
```

### Commands that currently support namespaces

You can omit the `setup.json` parameter to the following rtt commands if a default namespace is set:

- purge

## File structures

`rtt` handles two different configuration files, both written in json. One type is the `setup` file, that handles the rabbit topology and connection data.
You can find [an example here](examples/setup.json) and [the json schema file here](schemas/setup.schema.json).

The other type of file is the `rtt` file itself, where test messages and data are defined.
You can find multiple examples in this project: [with response message](examples/testA/testA.json) and [without response message](examples/testC.json).
The json schema is defined [here](schemas/rttfile.schema.json).

All those files are validated against the linked schema files before executing them, so the user can (hopefully) receive meaningful error messages in case the files are not valid.

> [!NOTE]
> The files are only validated **syntactically**. If (for example) an exchange is referenced in a queue that is not defined in the setup file, the app won't catch that before executing.

## Assertions

`rtt` supports assertions the user writes down in the rtt test file.
Let's talk about the options via [an example](examples/testD.json):
```json
"responseQueue": { // 1
    "queue": { // 2
        "exchange": "personal-info",
        "key": "structured",
        "name": "personal-info-structured"
    },
    "response": { // 3
        "identifier": { // 4
            "requestId": "3e4666bf-d5e5-4aa7-b8ce-cefe41c7568a"
        },
        "assertions": [ // 5
            { 
                "firstName": "John" // 6
            },
            {
                "lastName": "Doe" 
            },
            {
                "address": { // 7
                    "city": "Baltimore",
                    "street": "Main Street",
                    "zipcode": "21201"
                }
            },
            {
                "metaData.serviceData.serviceName": "rabbit-example-consumer" // 8
            }
        ]
    }
}
```

1. Top level element containing the queue information and the actual response assertions.
2. Queue information to tell `rtt` where to listen for the response. A `schema` element can be added for validation of the response.
3. Element containing the identifying field and the assertion to that specific message.
4. Object containing a single identifying field to uniquely identify a message in the queue. The value is not required to be a string but can be
a complex object. But the **key** must be **top level**.
5. List of assertions.
6. An example of an assertion of a top level field with a simple type (string, number, boolean)
7. An example of an assertion of a top level field with a complex type. **The asserted object must be exact to the incoming message to be evaluated to true**.
8. An example of a lower level field with a simple type. The accessor syntax to a lower level field is denoted via a dot.
**This means, you cannot match to JSON fields in a message that contain a dot themselves in the key.** Note that you can still match those fields when they are part of a complex object. Also, you are not limited to the comparison of a simple type. The expected value can be a complex object.

## Build

```sh
go get
go build
```

## Examples

### Use environment variables for username and password

In the `setup.json` you can use environment variables as substitutes for the credentials:
```json
{
    "connection": {
        "host": "localhost",
        "port": "5672",
        "user": "$(RABBIT_USER)",
        "password": "$(RABBIT_PASSWORD)"
    }
}
```

### Creating a queue with additional arguments

```json
{
    "exchange": "exampleExchange",
    "key": "test1",
    "name": "example1",
    "amqpTable": {
        "x-queue-type": "quorum"
    }
}
```