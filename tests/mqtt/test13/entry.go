package test13

import (
	"testing"

	"sync/atomic"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	testTypes "github.com/troian/surgemq/tests/types"
)

var topics []string
var wildTopics []string

type impl struct {
}

var _ testTypes.Provider = (*impl)(nil)

const (
	testName = "overlapping"
)

func init() {
	topics = []string{"TopicA", "TopicA/B", "Topic/C", "TopicA/C", "/TopicA"}
	wildTopics = []string{"TopicA/+", "+/C", "#", "/#", "/+", "+/+", "TopicA/#"}
}

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
	timeout := 5 * time.Second

	cfg := config.Get()

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("overlapping_test").
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

	mS := map[string]byte{
		wildTopics[6]: 2,
		wildTopics[0]: 1,
	}

	var count int32

	token = c.SubscribeMultiple(mS, func(client MQTT.Client, message MQTT.Message) {
		atomic.AddInt32(&count, 1)
	})

	token.Wait()
	require.NoError(t, token.Error())

	token = c.Publish(topics[3], 2, false, []byte("overlapping topic filters"))
	token.Wait()
	require.NoError(t, token.Error())

	<-time.After(timeout)

	c.Disconnect(1000)

	if atomic.LoadInt32(&count) == 2 {
		t.Log("This server is publishing one message for all matching overlapping subscriptions, not one for each")
	}
}
