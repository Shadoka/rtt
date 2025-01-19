# rtt - rabbit test tool

## Why and what

`rtt` is a test tool for rabbit queues aimed to be used similarly to tools like Postman and Bruno for HTTP calls.
Contrary to those tools `rtt` is purely a CLI tool.

Developers can create JSON files containing queue data (exchanges, routing keys, queue names) and message data.
For an example of a simple file [see this example file](examples/testB/testB.json). Optionally, you can define [a schema location](examples/schemas/example-schema.schema.json) to validate the message data contained in the file.

Additionally, if the consumer responds with a reply message it is possible to define the expected response message, [as you can see in this example](examples/testA/testA.json).

The general connection information (credentials & rabbit connection data) are defined in a file called [setup.json](examples/setup.json). This file can also define the whole topology of the rabbit infrastructure and should be referenced from each rtt test file. This has the advantage of not duplicating the credentials for each test and also to create the rabbit topology independent of the execution of a test file.

> [!NOTE]
> Currently `rtt` only supports a deep equals on the whole message body, which is not useful if the response message contains variable data like a current timestamp. It is planned to support an assert syntax to write more meaningful tests.

> [!NOTE]
> To match responses to messages we have sent via `rtt`, the tool currently assumes that the message id of the outgoing message is reflected into the message id of the incoming message. This is not realistic and will be changed.

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

## Build

```sh
go get
go build
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