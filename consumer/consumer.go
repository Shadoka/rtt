package consumer

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"rtt/data"
	"rtt/rabbit"
	"rtt/rttio"
	"slices"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ReplyConsumers is a map of queue name -> ResponseConsumer
var ReplyConsumers map[string]*ResponseConsumer

type ExpectedMessage struct {
	TestName    string
	MessageData data.Response
}

type ResponseConsumer struct {
	QueueInfo        data.RabbitQueue
	ExpectedMessages map[MessageId]ExpectedMessage
}

func Init() {
	ReplyConsumers = make(map[string]*ResponseConsumer)
}

func (consumer *ResponseConsumer) ListenForReplies(channel *amqp.Channel, notifyChannel chan data.ConsumerResult) {
	msgs := rabbit.CreateConsumer(channel, &consumer.QueueInfo)
	consumerResult := data.ConsumerResult{}
	consumerResult.ConsumerQueue = consumer.QueueInfo.Name

	for msg := range msgs {
		// deserialize response data
		var responseData map[string]interface{}
		if err := json.Unmarshal(msg.Body, &responseData); err != nil {
			consumerResult.AssertionError = err
			continue
		}

		// extract identifier value
		hasMatch, matchedMsgId := consumer.matchesAnyExpectedMessage(responseData)

		if hasMatch {
			expectedMessage := consumer.ExpectedMessages[*matchedMsgId]

			for _, assertion := range expectedMessage.MessageData.Assertions {
				if assertionErr := assert(assertion, responseData, expectedMessage.TestName); assertionErr != nil {
					consumerResult.AssertionError = assertionErr
				}
			}
			delete(consumer.ExpectedMessages, *matchedMsgId)
		} else {
			unexpectedMsgErr := fmt.Errorf("received an unexpected response message")
			consumerResult = data.ConsumerResult{
				AssertionError: unexpectedMsgErr,
			}
		}

		if len(consumer.ExpectedMessages) == 0 {
			break
		}
	}
	notifyChannel <- consumerResult
}

func AddResponse(msgId MessageId, consumerData data.ResponseQueue, testName string) {
	if ReplyConsumers[consumerData.Queue.Name] == nil {
		ReplyConsumers[consumerData.Queue.Name] = &ResponseConsumer{
			ExpectedMessages: map[MessageId]ExpectedMessage{
				msgId: ExpectedMessage{
					TestName:    testName,
					MessageData: consumerData.Response,
				},
			},
			QueueInfo: consumerData.Queue,
		}
	} else {
		ReplyConsumers[consumerData.Queue.Name].ExpectedMessages[msgId] = ExpectedMessage{
			TestName:    testName,
			MessageData: consumerData.Response,
		}
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

func assert(assertion map[string]interface{}, responseData map[string]interface{}, testName string) error {
	assertionMessage := createAssertionMessage(assertion, testName)
	assertionResult := assertExpectationToResponse(assertion, responseData)

	if assertionResult != SUCCESS {
		assertionErr := fmt.Errorf("assertion error when comparing received message with expected message")
		rttio.PrintAssertionResult(assertionMessage, false)
		return assertionErr
	} else {
		rttio.PrintAssertionResult(assertionMessage, true)
	}
	return nil
}

func createAssertionMessage(assertion map[string]interface{}, testName string) string {
	keys := slices.Collect(maps.Keys(assertion))
	value := assertion[keys[0]]
	return fmt.Sprintf("(%v) %v: %v", testName, keys[0], value)
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
	value, found := findField(fieldName[0], actualData)
	if found {
		if reflect.DeepEqual(value, expectation[fieldName[0]]) {
			return SUCCESS
		} else {
			return MISMATCHED_EXPECTATION
		}
	}
	return NO_SUCH_FIELD
}

func findField(expectedFieldName string, actualData map[string]interface{}) (interface{}, bool) {
	// TODO: Constant for "."
	nestedKeys := strings.Split(expectedFieldName, ".")
	var fieldValue interface{}
	var found bool
	if len(nestedKeys) == 1 {
		fieldValue, found = actualData[expectedFieldName]
	} else {
		fieldValue, found = findNestedField(nestedKeys, actualData)
	}
	return fieldValue, found
}

func findNestedField(nestedKeys []string, messageData map[string]interface{}) (interface{}, bool) {
	var currentSelection map[string]interface{} = messageData
	lastKeyIndex := len(nestedKeys) - 1
	for i, v := range nestedKeys {
		if i != lastKeyIndex {
			// navigate down the json path
			currentNavigationObject := currentSelection[v]
			if currentNavigationObject == nil {
				return nil, false
			} else {
				currentSelection = currentNavigationObject.(map[string]interface{})
			}
		} else {
			// we are at the level we expect to find the test data
			var fieldValue interface{}
			var found bool
			fieldValue, found = currentSelection[v]
			return fieldValue, found
		}
	}

	return nil, false
}

type AssertionResult int

const (
	SUCCESS AssertionResult = iota
	NO_SUCH_FIELD
	MISMATCHED_EXPECTATION
)

type MessageId struct {
	IdName  string
	IdValue string
}
