package test4

import (
	assert "github.com/stretchr/testify/assert"
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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
	payload := []byte("a much longer message that we can shorten to the extent that we need to")
	test_topic := "async test topic"

	cfg := config.Get()

	var wg0 sync.WaitGroup
	var wg1 sync.WaitGroup

	wg0.Add(1)
	wg1.Add(1)

	defaultPublishHandler := func(_ MQTT.Client, msg MQTT.Message) {
		recv := msg.Payload()[:len(payload)]
		if assert.Equal(t, payload, recv, "Invalid payload for QoS %d. expected/received [%s]/[%s]", msg.Qos(), string(payload), string(recv)) {
			defer wg1.Done()
		}
	}

	var c MQTT.Client

	onConnect := func(cl MQTT.Client) {
		defer wg0.Done()

		token := c.Subscribe(test_topic, 1, nil)
		token.Wait()
		assert.NoError(t, token.Error())

		token = cl.Publish(test_topic, 1, false, payload)
		token.Wait()
		assert.NoError(t, token.Error())
	}

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("async_test").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		SetWill("will topic", "will message", 1, false).
		SetOnConnectHandler(onConnect).
		SetDefaultPublishHandler(defaultPublishHandler)

	c = MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	if assert.Equal(t, false, waitTimeout(&wg0, 5*time.Second), "Error waiting connect") {
		assert.Equal(t, false, waitTimeout(&wg1, 5*time.Second), "Error waiting message")

		token = c.Unsubscribe(test_topic)
		token.Wait()
		require.NoError(t, token.Error())

		c.Disconnect(250)
	}
}
