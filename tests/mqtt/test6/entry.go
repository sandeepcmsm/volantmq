package test6

import (
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	//"sync"
	//"time"
	"bytes"
	"time"

	"github.com/troian/surgemq/tests/mqtt/config"
	testTypes "github.com/troian/surgemq/tests/types"
)

type impl struct {
}

var _ testTypes.Provider = (*impl)(nil)

const (
	testName = "connection lost and will messages"
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
	will_topic := "GO Test 6: will topic"
	will_message := "will message from Client-1"

	var c MQTT.Client
	var c2 MQTT.Client

	cfg := config.Get()

	lost := make(chan bool)
	willArrived := make(chan bool)

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.ProxyHost).
		SetClientID("Client_1").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(2).
		SetWill(will_topic, will_message, 2, false).
		//SetDefaultPublishHandler(defaultPublishHandler).
		SetConnectionLostHandler(func(client MQTT.Client, reason error) {
			lost <- true
			client.Disconnect(0)
		})

	opts1 := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("Client_2").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		//SetDefaultPublishHandler(defaultPublishHandler).
		SetConnectionLostHandler(func(client MQTT.Client, reason error) {
			assert.Fail(t, reason.Error())
		})

	// Client-1 with Will options
	c = MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	// Client - 2 (multi-threaded)
	c2 = MQTT.NewClient(opts1)
	token = c2.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	token = c2.Subscribe(will_topic, 2, func(client MQTT.Client, msg MQTT.Message) {
		if msg.Topic() == will_topic && bytes.Equal([]byte(will_message), msg.Payload()) {
			willArrived <- true
		}
	})
	token.Wait()
	require.NoError(t, token.Error())

	c.Publish(cfg.MqttSasTopic, 0, false, "TERMINATE")

	waitChan := func(ch chan bool, timeout time.Duration) bool {
		select {
		case <-ch:
			return false
		case <-time.After(5 * time.Second):
			return true
		}
	}

	assert.Equal(t, false, waitChan(lost, 5*time.Second), "Timeout waiting connection lost event")
	assert.Equal(t, false, waitChan(willArrived, 5*time.Second), "Timeout waiting will message")

	//token = c2.Unsubscribe(will_topic)
	//token.Wait()
	//require.NoError(t, token.Error())

	require.Equal(t, true, c2.IsConnected(), "Client-2 should stay connected")
	require.Equal(t, false, c.IsConnected(), "Client-1 should be disconnected")

	c2.Disconnect(250)
}
