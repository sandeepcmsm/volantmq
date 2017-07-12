package test5

import (
	"testing"

	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	testTypes "github.com/troian/surgemq/tests/types"
)

type impl struct {
}

var _ testTypes.Provider = (*impl)(nil)

const (
	testName = "disconnect with quiesce timeout should allow exchanges to complete"
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
	test_topic := "Persistence test 2"
	subsQos := byte(2)

	cfg := config.Get()

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("xrctest1_test_5").
		SetCleanSession(false).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		SetConnectionLostHandler(func(client MQTT.Client, reason error) {
			t.Logf("MQTT lost connection: %s", reason.Error())
		})

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	token = c.Subscribe(test_topic, subsQos, nil)
	token.Wait()
	require.NoError(t, token.Error())

	tokens := []MQTT.Token{}

	for i := 0; i < 3; i++ {
		payload := fmt.Sprintf("Message sequence no %d", i)
		tok := c.Publish(test_topic, 1, false, payload)
		tokens = append(tokens, tok)
	}

	for i, tok := range tokens {
		assert.Equal(t, true, tok.WaitTimeout(10*time.Second), "Error waiting token %d", i)
		assert.NoError(t, tok.Error())
	}

	c.Disconnect(0)

	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	token = c.Unsubscribe(test_topic)
	require.NoError(t, token.Error())

	c.Disconnect(250)
}
