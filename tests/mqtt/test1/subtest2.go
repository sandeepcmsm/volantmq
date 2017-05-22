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

func SubTest2(t *testing.T) {
	test_topic := "GO client test2"
	subsQos := byte(2)
	payload := []byte("a much longer message that we can shorten to the extent that we need to")
	var failures int
	received := make(chan MQTT.Message)

	defaultPublishHandler := func(_ MQTT.Client, msg MQTT.Message) {
		received <- msg

		failed := false

		if !assert.Equal(t, string(payload), string(msg.Payload()), "Received unexpected payload") {
			failed = true
		}

		if failed {
			failures++
		}
	}

	cfg := config.Get()

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("multi_threaded_test").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		SetDefaultPublishHandler(defaultPublishHandler).
		SetConnectionLostHandler(func(client MQTT.Client, reason error) {
			t.Logf("MQTT lost connection: %s", reason.Error())
		})

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	token = c.Subscribe(test_topic, subsQos, func(client MQTT.Client, msg MQTT.Message) {
		defaultPublishHandler(client, msg)
	})

	token.Wait()
	require.NoError(t, token.Error())

	subTest2SendReceive(received, t, c, 0, test_topic, payload)
	subTest2SendReceive(received, t, c, 1, test_topic, payload)
	subTest2SendReceive(received, t, c, 2, test_topic, payload)

	token = c.Unsubscribe(test_topic)
	require.NoError(t, token.Error())

	require.Equal(t, 0, failures, "There are unmatching messages received")

	c.Disconnect(250)

	require.Equal(t, 0, failures, "Failed messages")
}

func subTest2SendReceive(r chan MQTT.Message, t *testing.T, c MQTT.Client, qos byte, topic string, payload []byte) {
	iterations := 50

	var wgDelivered sync.WaitGroup

	wgDelivered.Add(iterations)

	for i := 0; i < iterations; i++ {
		token := c.Publish(topic, qos, false, payload)
		require.NoError(t, token.Error())

		go func(tok MQTT.Token) {
			tok.Wait()
			if assert.NoError(t, token.Error()) {
				wgDelivered.Done()
			}
		}(token)

		// wait message has arrived
		var timeout bool
		select {
		case <-r:
			timeout = false // completed normally
		case <-time.After(10 * time.Second):
			timeout = true // timed out
		}
		require.Equal(t, false, timeout, "Timed out waiting for message")
	}

	if qos > 0 {
		res := waitTimeout(&wgDelivered, 10*time.Second)
		require.Equal(t, false, res, "Timed out waiting for deliveries")
	}
}
