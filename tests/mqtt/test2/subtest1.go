package test2

import (
	"testing"

	assert "github.com/stretchr/testify/assert"

	"sync"
	"sync/atomic"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/troian/surgemq/tests/mqtt/config"
)

type subTest1Config struct {
	wg         *sync.WaitGroup
	done       *sync.WaitGroup
	payload    []byte
	topic      string
	qos        byte
	iterations int
	timeout    time.Duration
}

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
	var failures int32
	iterations := 50
	payload := []byte("a much longer message that we can shorten to the extent that we need to")
	test_topic := "GO client test1"

	cfg := config.Get()

	var subDone sync.WaitGroup
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
		SetClientID("single_object, multiple threads").
		SetCleanSession(true).
		SetUsername(cfg.TestUser).
		SetPassword(cfg.TestPassword).
		SetAutoReconnect(false).
		SetKeepAlive(20).
		SetDefaultPublishHandler(defaultPublishHandler)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	subDone.Add(3)

	go subTest1SendReceive(t, c, subTest1Config{
		wg:         &wg0,
		done:       &subDone,
		payload:    payload,
		topic:      test_topic,
		qos:        0,
		iterations: iterations,
		timeout:    10,
	})
	go subTest1SendReceive(t, c, subTest1Config{
		wg:         &wg1,
		done:       &subDone,
		payload:    payload,
		topic:      test_topic,
		qos:        1,
		iterations: iterations,
		timeout:    10,
	})
	go subTest1SendReceive(t, c, subTest1Config{
		wg:         &wg2,
		done:       &subDone,
		payload:    payload,
		topic:      test_topic,
		qos:        2,
		iterations: iterations,
		timeout:    10,
	})

	assert.Equal(t, false, waitTimeout(&subDone, 30*time.Second))

	require.Equal(t, int32(0), failures, "Messages failed %d", failures)

	token = c.Unsubscribe(test_topic)
	token.Wait()
	require.NoError(t, token.Error())

	c.Disconnect(250)

	// Just to make sure we can connect again
	c = MQTT.NewClient(opts)
	token = c.Connect()
	token.Wait()
	require.NoError(t, token.Error())
	c.Disconnect(250)
}

func subTest1SendReceive(t *testing.T, c MQTT.Client, cfg subTest1Config) {
	defer cfg.done.Done()

	c.Subscribe(cfg.topic, cfg.qos, nil)

	for i := 0; i < cfg.iterations; i++ {
		c.Publish(cfg.topic, cfg.qos, false, cfg.payload)
	}

	assert.Equal(t, false, waitTimeout(cfg.wg, cfg.timeout*time.Second))
}
