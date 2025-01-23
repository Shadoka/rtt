package consumer

import (
	"encoding/json"
	"fmt"
	"maps"
	"rtt/data"
	"rtt/rabbit"
	"slices"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ReplyConsumers is a map of queue name -> ResponseConsumer
var ReplyConsumers map[string]*ResponseConsumer

type ResponseConsumer struct {
	QueueInfo        data.RabbitQueue
	ExpectedMessages map[MessageId]data.Response
}

func Init() {
	ReplyConsumers = make(map[string]*ResponseConsumer)
}

func (consumer *ResponseConsumer) ListenForReplies(channel *amqp.Channel, notifyChannel chan data.ApplicationResult) {
	msgs := rabbit.CreateConsumer(channel, &consumer.QueueInfo)

	for msg := range msgs {
		appResult := data.ApplicationResult{}

		// deserialize response data
		var responseData map[string]interface{}
		if err := json.Unmarshal(msg.Body, &responseData); err != nil {
			appResult.AssertionError = err
			notifyChannel <- appResult
			continue
		}

		// extract identifier value
		hasMatch, matchedMsgId := consumer.matchesAnyExpectedMessage(responseData)

		if hasMatch {
			expectedMessage := consumer.ExpectedMessages[*matchedMsgId]

			for _, assertion := range expectedMessage.Assertions {
				assertionResult := assertExpectationToResponse(assertion, responseData)

				if assertionResult != SUCCESS {
					assertionErr := fmt.Errorf("assertion error when comparing received message with expected message")
					appResult.AssertionError = assertionErr
				}
			}
			delete(consumer.ExpectedMessages, *matchedMsgId)
		} else {
			unexpectedMsgErr := fmt.Errorf("received an unexpected response message")
			appResult = data.ApplicationResult{AssertionError: unexpectedMsgErr}
		}

		notifyChannel <- appResult
		if len(consumer.ExpectedMessages) == 0 {
			break
		}
	}
}

func AddResponse(msgId MessageId, consumerData data.ResponseQueue) {
	if ReplyConsumers[consumerData.Queue.Name] == nil {
		ReplyConsumers[consumerData.Queue.Name] = &ResponseConsumer{
			ExpectedMessages: map[MessageId]data.Response{
				msgId: consumerData.Response,
			},
			QueueInfo: consumerData.Queue,
		}
	} else {
		ReplyConsumers[consumerData.Queue.Name].ExpectedMessages[msgId] = consumerData.Response
	}
}

func CreateMessageIdFromResponse(response *data.Response) MessageId {
	var msgId MessageId = MessageId{}

	for name, value := range response.Identifier {
		msgId.IdName = name
		msgId.IdValue = fmt.Sprintf("%v", value)
	}

	return msgId
}

type MessageId struct {
	IdName  string
	IdValue string
}

func GetMaxReplyCount() int {
	result := 0

	for _, consumer := range ReplyConsumers {
		result += len(consumer.ExpectedMessages)
	}

	return result
}

func printMap(currentMap map[string]interface{}) {
	for k, v := range currentMap {
		fmt.Printf("key: %v, value: %v\n", k, v)
	}
}

func (consumer *ResponseConsumer) matchesAnyExpectedMessage(messageData map[string]interface{}) (bool, *MessageId) {
	for expId, _ := range consumer.ExpectedMessages {
		if messageIdValue, found := messageData[expId.IdName]; found {
			if messageIdValue == expId.IdValue {
				return true, &expId
			}
		}
	}
	return false, nil
}

func assertExpectationToResponse(expectation map[string]interface{}, actualData map[string]interface{}) AssertionResult {
	// TODO: print out information about failures
	fieldName := slices.Collect(maps.Keys(expectation))
	value, found := actualData[fieldName[0]]
	if found {
		if value == expectation[fieldName[0]] {
			return SUCCESS
		} else {
			return MISMATCHED_EXPECTATION
		}
	}
	return NO_SUCH_FIELD
}

type AssertionResult int

const (
	SUCCESS AssertionResult = iota
	NO_SUCH_FIELD
	MISMATCHED_EXPECTATION
)
