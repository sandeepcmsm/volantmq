package test12

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
	testName = "offline queuing"
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
		SetClientID("offline_test1").
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
	c.Disconnect(0)

	opts.SetCleanSession(false)
	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	retainCount := 3

	var wg sync.WaitGroup
	wg.Add(retainCount)

	token = c.Subscribe(wildTopics[5], 2, nil)
	token.Wait()
	require.NoError(t, token.Error())

	c.Disconnect(0)

	opts.SetClientID("offline_test2")
	opts.SetCleanSession(true)
	c2 := MQTT.NewClient(opts)
	token = c2.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	for i := byte(0); i < byte(retainCount); i++ {
		token = c2.Publish(topics[i+1], i, true, "qos "+strconv.Itoa(int(i)))
		token.Wait()
		require.NoError(t, token.Error())
	}
	c2.Disconnect(0)

	opts.SetClientID("offline_test1")
	opts.SetCleanSession(false)
	opts.SetDefaultPublishHandler(func(client MQTT.Client, message MQTT.Message) {
		wg.Done()
	})

	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())
	assert.Equal(t, false, testTypes.WaitTimeout(&wg, timeout), "Timeout waiting retained messages")

	c.Disconnect(0)

	// Clean session
	opts.SetCleanSession(true)
	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())
	c.Disconnect(0)
}
