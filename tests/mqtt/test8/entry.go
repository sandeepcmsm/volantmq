package test8

import (
	assert "github.com/stretchr/testify/assert"
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
	testTypes "github.com/troian/surgemq/tests/types"
	"sync"
	"sync/atomic"
	"time"
)

type impl struct {
}

var _ testTypes.Provider = (*impl)(nil)

const (
	testName = "multiple threads using callbacks"
)

func New() testTypes.Provider {
	return &impl{}
}

func (im *impl) Name() string {
	return testName
}

func (im *impl) Run(t *testing.T) {
	var failures int32
	iterations := 50
	payload := []byte("a much longer message that we can shorten to the extent that we need to")
	test_topic := "GO client test2"
	subsqos := byte(2)

	cfg := config.Get()

	var wg0 sync.WaitGroup
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	wg0.Add(iterations)
	wg1.Add(iterations)
	wg2.Add(iterations)

	defaultPublishHandler := func(_ MQTT.Client, msg MQTT.Message) {
		recv := msg.Payload()[:len(payload)]
		if !assert.Equal(t, payload, recv, "Invalid payload for QoS %d. expected/received [%s]/[%s]", msg.Qos(), string(payload), string(recv)) {
			atomic.AddInt32(&failures, 1)
		} else {
			switch msg.Qos() {
			case 0:
				wg0.Done()
			case 1:
				wg1.Done()
			case 2:
				wg2.Done()
			}
		}
	}

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Host).
		SetClientID("multi_threaded_sample").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		SetDefaultPublishHandler(defaultPublishHandler).
		SetConnectionLostHandler(func(client MQTT.Client, reason error) {
			assert.Fail(t, reason.Error())
		})

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	token = c.Subscribe(test_topic, subsqos, nil)
	token.Wait()
	require.NoError(t, token.Error())

	type subTestConfig struct {
		wg         *sync.WaitGroup
		done       *sync.WaitGroup
		payload    []byte
		topic      string
		qos        byte
		iterations int
		timeout    time.Duration
	}

	worker := func(cfg subTestConfig) {
		for i := 0; i < cfg.iterations; i++ {
			c.Publish(cfg.topic, cfg.qos, false, cfg.payload)
		}

		assert.Equal(t, false, testTypes.WaitTimeout(cfg.wg, cfg.timeout*time.Second))
	}

	worker(subTestConfig{
		wg:         &wg0,
		payload:    payload,
		topic:      test_topic,
		qos:        0,
		iterations: iterations,
		timeout:    30,
	})
	worker(subTestConfig{
		wg:         &wg1,
		payload:    payload,
		topic:      test_topic,
		qos:        1,
		iterations: iterations,
		timeout:    10,
	})
	worker(subTestConfig{
		wg:         &wg2,
		payload:    payload,
		topic:      test_topic,
		qos:        2,
		iterations: iterations,
		timeout:    10,
	})

	require.Equal(t, int32(0), failures, "Messages failed %d", failures)
	c.Disconnect(250)

	// Just to make sure we can connect again
	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())
	c.Disconnect(250)
}
