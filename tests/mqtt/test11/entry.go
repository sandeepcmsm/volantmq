package test11

import (
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	testTypes "github.com/troian/surgemq/tests/types"
	"strconv"
	"sync"
	"time"
)

var topics []string
var wildTopics []string

type impl struct {
}

var _ testTypes.Provider = (*impl)(nil)

const (
	testName = "retained messages"
)

func init() {
	topics = []string{"TopicA", "TopicA/B", "Topic/C", "TopicA/C", "/TopicA"}
	wildTopics = []string{"TopicA/+", "+/C", "#", "/#", "/+", "+/+", "TopicA/#"}
}

func New() testTypes.Provider {
	return &impl{}
}

func (im *impl) Name() string {
	return testName
}

func (im *impl) Run(t *testing.T) {
	timeout := 5 * time.Second

	cfg := config.Get()

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("retained_test").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		SetConnectionLostHandler(func(client MQTT.Client, reason error) {
			assert.Fail(t, reason.Error())
		})

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	retainCount := 3

	var wg sync.WaitGroup
	wg.Add(retainCount)

	for i := byte(0); i < byte(retainCount); i++ {
		token = c.Publish(topics[i+1], i, true, "qos "+strconv.Itoa(int(i)))
		token.Wait()
		require.NoError(t, token.Error())
	}

	token = c.Subscribe(wildTopics[5], 2, func(_ MQTT.Client, msg MQTT.Message) {
		wg.Done()
	})
	token.Wait()
	require.NoError(t, token.Error())

	assert.Equal(t, false, testTypes.WaitTimeout(&wg, timeout), "Timeout waiting retained messages")

	c.Disconnect(250)

	// Just to make sure we can connect again
	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	for i := byte(0); i < byte(retainCount); i++ {
		payload := []byte{}
		token = c.Publish(topics[i+1], i, true, payload)
		token.Wait()
		require.NoError(t, token.Error())
	}

	noRetains := 0
	token = c.Subscribe(wildTopics[5], 2, func(_ MQTT.Client, msg MQTT.Message) {
		noRetains++
	})
	token.Wait()
	require.NoError(t, token.Error())

	<-time.After(timeout)
	c.Disconnect(250)

	require.Equal(t, 0, noRetains, "No retained messages should be received after delete")
}
