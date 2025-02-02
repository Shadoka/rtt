# rtt - rabbit test tool

## Why and what

`rtt` is a test tool for rabbit queues aimed to be used similarly to tools like Postman and Bruno for HTTP calls.
Contrary to those tools `rtt` is purely a CLI tool.
Also, `rtt` can create topologies that are predefined in `setup.json` files.

Developers can create JSON files containing queue data (exchanges, routing keys, queue names) and message data.
For an example of a simple file [see this example file](examples/testB/testB.json). Optionally, you can define [a schema location](examples/schemas/example-schema.schema.json) to validate the message data contained in the file.

Additionally, if the consumer responds with a reply message it is possible to define the expected response message, [as you can see in this example](examples/testA/testA.json).

The general connection information (credentials & rabbit connection data) are defined in a file called [setup.json](examples/setup.json). This file can also define the whole topology of the rabbit infrastructure and should be referenced from each rtt test file. This has the advantage of not duplicating the credentials for each test and also to create the rabbit topology independent of the execution of a test file.

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