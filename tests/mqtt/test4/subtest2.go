package test4

import (
	//assert "github.com/stretchr/testify/assert"
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	"time"
)

func SubTest2(t *testing.T) {
	//payload := []byte("a much longer message that we can shorten to the extent that we need to")
	//test_topic := "async test topic"

	cfg := config.Get()

	//var wg0 sync.WaitGroup
	//var wg1 sync.WaitGroup
	//
	//wg0.Add(1)
	//wg1.Add(1)
	//
	//defaultPublishHandler := func(_ MQTT.Client, msg MQTT.Message) {
	//	recv := msg.Payload()[:len(payload)]
	//	if assert.Equal(t, payload, recv, "Invalid payload for QoS %d. expected/received [%s]/[%s]", msg.Qos(), string(payload), string(recv)) {
	//		defer wg1.Done()
	//	}
	//}
	//
	//
	//
	//onConnect := func(cl MQTT.Client) {
	//	defer wg0.Done()
	//
	//	token := c.Subscribe(test_topic, 1, nil)
	//	token.Wait()
	//	assert.NoError(t, token.Error())
	//
	//	token = cl.Publish(test_topic, 1, false, payload)
	//	token.Wait()
	//	assert.NoError(t, token.Error())
	//}
	var c MQTT.Client

	opts := MQTT.NewClientOptions().
		AddBroker("tcp://9.20.96.160:66").
		SetClientID("connect timeout").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetConnectTimeout(5*time.Second).
		SetKeepAlive(20).
		SetWill("will topic", "will message", 1, false)

	c = MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.EqualError(t, nil, token.Error().Error())
}
