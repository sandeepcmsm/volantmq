package test1

import (
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	MQTTPackets "github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
)

func SubTest3(t *testing.T) {
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
	require.Equal(t, true, token.Wait())
	require.NoError(t, token.Error())
	c.Disconnect(250)

	opts.SetUsername("badusername")
	c = MQTT.NewClient(opts)
	token = c.Connect()
	require.Equal(t, true, token.Wait())
	require.EqualError(t, MQTTPackets.ConnErrors[MQTTPackets.ErrRefusedBadUsernameOrPassword], token.Error().Error())

	opts.SetUsername(cfg.TestUser)
	opts.SetClientID("")

	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())
	c.Disconnect(250)

	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetCleanSession(false)
	c = MQTT.NewClient(opts)
	token = c.Connect()
	require.Equal(t, true, token.Wait())
	require.EqualError(t, MQTTPackets.ConnErrors[MQTTPackets.ErrRefusedIDRejected], token.Error().Error())
}
