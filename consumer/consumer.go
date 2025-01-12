package consumer

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"rtt/data"
	"rtt/rabbit"
	"slices"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ReplyConsumers is a map of queue name -> ResponseConsumer
var ReplyConsumers map[string]*ResponseConsumer

type ResponseConsumer struct {
	QueueInfo        data.RabbitQueue
	ExpectedMessages map[string]data.Response
}

func Init() {
	ReplyConsumers = make(map[string]*ResponseConsumer)
}

func (consumer *ResponseConsumer) ListenForReplies(channel *amqp.Channel, notifyChannel chan data.ApplicationResult) {
	msgs := rabbit.CreateConsumer(channel, &consumer.QueueInfo)

	expectedIds := slices.Collect(maps.Keys(consumer.ExpectedMessages))
	msgProcessingCount := 0
	for msg := range msgs {
		currentExpectedIds := consumer.getUpdatedListOfRemainingMessages(expectedIds, msg.MessageId)

		if len(currentExpectedIds) < len(expectedIds) {
			expectedIds = currentExpectedIds
			msgProcessingCount++
			appResult := data.ApplicationResult{}

			var responseData map[string]interface{}
			if err := json.Unmarshal(msg.Body, &responseData); err != nil {
				appResult.AssertionError = err
				notifyChannel <- appResult
				continue
			}

			expectedMessage := consumer.ExpectedMessages[msg.MessageId]
			// TODO: Write own comparison with detailed assertion mismatch information
			if reflect.DeepEqual(expectedMessage.Data, responseData) {
				notifyChannel <- appResult
			} else {
				assertionErr := fmt.Errorf("assertion error when comparing received message with expected message")
				appResult.AssertionError = assertionErr
				notifyChannel <- appResult
			}
		} else {
			unexpectedMsgErr := fmt.Errorf("received an unexpected response message")
			appResult := data.ApplicationResult{AssertionError: unexpectedMsgErr}
			notifyChannel <- appResult
		}

		if msgProcessingCount == len(consumer.ExpectedMessages) {
			break
		}
	}
}

func (consumer *ResponseConsumer) getUpdatedListOfRemainingMessages(remainingMessageIds []string, incomingMessageId string) []string {
	// NOTE: If the application has a reply queue of its own it may not be possible to 'reflect' the message id
	return slices.DeleteFunc(remainingMessageIds, func(expectedMessageId string) bool {
		return incomingMessageId == expectedMessageId
	})
}

func printMap(currentMap map[string]interface{}) {
	for k, v := range currentMap {
		fmt.Printf("key: %v, value: %v\n", k, v)
	}
}

func AddResponse(msgId string, consumerData data.ResponseQueue) {
	if ReplyConsumers[consumerData.Queue.Name] == nil {
		ReplyConsumers[consumerData.Queue.Name] = &ResponseConsumer{
			ExpectedMessages: map[string]data.Response{
				msgId: consumerData.Response,
			},
			QueueInfo: consumerData.Queue,
		}
	} else {
		ReplyConsumers[consumerData.Queue.Name].ExpectedMessages[msgId] = consumerData.Response
	}
}

func GetMaxReplyCount() int {
	result := 0

	for _, consumer := range ReplyConsumers {
		result += len(consumer.ExpectedMessages)
	}

	return result
}
