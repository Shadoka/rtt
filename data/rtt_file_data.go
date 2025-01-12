package data

import "encoding/json"

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
}

// RABBIT STRUCTURES

type Connection struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type RabbitExchange struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"autoDelete"`
	Exclusive  bool   `json:"exclusive"`
	NoWait     bool   `json:"noWait"`
	Internal   bool   `json:"internal"`
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
	}

	if err := json.Unmarshal(text, &opts); err != nil {
		return err
	}

	*re = RabbitExchange(opts)
	return nil
}

type RabbitQueue struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Exchange   string `json:"exchange"`
	Schema     string `json:"schema"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"autoDelete"`
	Exclusive  bool   `json:"exclusive"`
	NoWait     bool   `json:"noWait"`
}

func (rq *RabbitQueue) UnmarshalJSON(text []byte) error {
	type defaults RabbitQueue

	opts := defaults{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
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
	//Status  string                 `json:"status"`
	//Message string                 `json:"message"`
	Data map[string]interface{} `json:"data"`
}
