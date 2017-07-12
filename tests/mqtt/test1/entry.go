package test1

import (
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	testTypes "github.com/troian/surgemq/tests/types"
)

type impl struct {
}

var _ testTypes.Provider = (*impl)(nil)

const (
	testName = "single threaded client using receive"
)

// nolint: golint
func New() testTypes.Provider {
	return &impl{}
}

// nolint: golint
func (im *impl) Name() string {
	return testName
}

// nolint: golint
func (im *impl) Run(t *testing.T) {
	test_topic := "GO client test1"
	subsQos := byte(2)

	cfg := config.Get()

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("single_threaded_test").
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

	receiver := make(chan MQTT.Message)

	token = c.Subscribe(test_topic, subsQos, func(_ MQTT.Client, msg MQTT.Message) {
		receiver <- msg
	})

	token.Wait()
	require.NoError(t, token.Error())

	worker := func(qos byte) {
		iterations := 50
		payload := "a much longer message that we can shorten to the extent that we need to payload up to 11"

		for i := 0; i < iterations; i++ {
			lTok := c.Publish(test_topic, qos, false, payload)
			lTok.Wait()
			require.NoError(t, token.Error())

			msg := <-receiver

			require.Equal(t, test_topic, msg.Topic())
			require.Equal(t, qos, msg.Qos())
			require.Equal(t, payload, string(msg.Payload()))
		}
	}

	worker(0)
	worker(1)
	worker(2)

	token = c.Unsubscribe(test_topic)
	require.NoError(t, token.Error())

	c.Disconnect(250)

	close(receiver)

	// Just to make sure we can connect again
	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())
	c.Disconnect(250)
}
