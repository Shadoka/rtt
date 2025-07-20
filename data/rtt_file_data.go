package data

import (
	"encoding/json"
	"os"
	"rtt/constants"
	"slices"
	"strings"
)

// FILE STRUCTURES

type RttFile struct {
	Name           string        `json:"name"`
	ConnectionFile string        `json:"connectionFile"`
	InputQueue     InputQueue    `json:"inputQueue"`
	ResponseQueue  ResponseQueue `json:"responseQueue"`
}

type SetupFile struct {
	Exchanges  []RabbitExchange `json:"exchanges"`
	Queues     []RabbitQueue    `json:"queues"`
	Connection Connection       `json:"connection"`
	Protected  bool             `json:"protected"`
}

func (namespace *SetupFile) ContainsQueue(queueName string) bool {
	return slices.ContainsFunc(namespace.Queues, func(q RabbitQueue) bool {
		return q.Name == queueName
	})
}

func (namespace *SetupFile) GetQueueInfo(queueName string) *RabbitQueue {
	for _, rq := range namespace.Queues {
		if rq.Name == queueName {
			return &rq
		}
	}
	return nil
}

type ConfigFile struct {
	DefaultNamespaceSetup string `json:"defaultNamespaceSetup"`
	DefaultNamespaceAlias string `json:"defaultNamespaceAlias"`
}

// RABBIT STRUCTURES

type Connection struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (c *Connection) UnmarshalJSON(text []byte) error {
	type dummy Connection
	d := dummy{}

	if err := json.Unmarshal(text, &d); err != nil {
		return err
	}

	if strings.HasPrefix(d.User, constants.ENV_VAR_PREFIX) {
		d.User = GetEnvVar(d.User)
	}

	if strings.HasPrefix(d.Password, constants.ENV_VAR_PREFIX) {
		d.Password = GetEnvVar(d.Password)
	}

	*c = Connection(d)
	return nil
}

type RabbitExchange struct {
	Name       string                 `json:"name"`
	Kind       string                 `json:"kind"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"autoDelete"`
	Exclusive  bool                   `json:"exclusive"`
	NoWait     bool                   `json:"noWait"`
	Internal   bool                   `json:"internal"`
	AmqpTable  map[string]interface{} `json:"amqpTable"`
}

func (re *RabbitExchange) UnmarshalJSON(text []byte) error {
	type defaults RabbitExchange

	opts := defaults{
		Kind:       "direct",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Internal:   false,
		AmqpTable:  nil,
	}

	if err := json.Unmarshal(text, &opts); err != nil {
		return err
	}

	*re = RabbitExchange(opts)
	return nil
}

type RabbitQueue struct {
	Name       string                 `json:"name"`
	Key        string                 `json:"key"`
	Exchange   string                 `json:"exchange"`
	Schema     string                 `json:"schema"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"autoDelete"`
	Exclusive  bool                   `json:"exclusive"`
	NoWait     bool                   `json:"noWait"`
	AmqpTable  map[string]interface{} `json:"amqpTable"`
}

func (rq *RabbitQueue) UnmarshalJSON(text []byte) error {
	type defaults RabbitQueue

	opts := defaults{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		AmqpTable:  nil,
	}

	if err := json.Unmarshal(text, &opts); err != nil {
		return err
	}

	*rq = RabbitQueue(opts)
	return nil
}

// TEST STRUCTURES

type InputQueue struct {
	Queue RabbitQueue            `json:"queue"`
	Data  map[string]interface{} `json:"data"`
}

type ResponseQueue struct {
	Queue    RabbitQueue `json:"queue"`
	Response Response    `json:"response"`
}

type Response struct {
	Identifier map[string]interface{}   `json:"identifier"`
	Assertions []map[string]interface{} `json:"assertions"`
}

func (r *Response) UnmarshalJSON(text []byte) error {
	type defaults Response

	opts := defaults{
		Assertions: make([]map[string]interface{}, 0),
	}

	if err := json.Unmarshal(text, &opts); err != nil {
		return err
	}

	*r = Response(opts)
	return nil
}

// util - cant go into rttio because of cycles

func GetEnvVar(rawEnvVar string) string {
	noPrefix, _ := strings.CutPrefix(rawEnvVar, constants.ENV_VAR_PREFIX)
	noSuffix, _ := strings.CutSuffix(noPrefix, constants.ENV_VAR_SUFFIX)
	return os.Getenv(noSuffix)
}
