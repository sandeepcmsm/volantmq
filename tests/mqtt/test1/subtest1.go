package test1

import (
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	"sync"
	"time"
)

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func SubTest1(t *testing.T) {
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

	subTest1SendReceive(receiver, t, c, 0, test_topic)
	subTest1SendReceive(receiver, t, c, 1, test_topic)
	subTest1SendReceive(receiver, t, c, 2, test_topic)

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

func subTest1SendReceive(r chan MQTT.Message, t *testing.T, c MQTT.Client, qos byte, topic string) {
	iterations := 50
	payload := "a much longer message that we can shorten to the extent that we need to payload up to 11"

	for i := 0; i < iterations; i++ {
		token := c.Publish(topic, qos, false, payload)
		token.Wait()
		require.NoError(t, token.Error())

		msg := <-r

		require.Equal(t, topic, msg.Topic())
		require.Equal(t, qos, msg.Qos())
		require.Equal(t, payload, string(msg.Payload()))
	}
}
