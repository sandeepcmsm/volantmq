package tests

import (
	"github.com/troian/surgemq/auth"
	authTypes "github.com/troian/surgemq/auth/types"
	persistType "github.com/troian/surgemq/persistence/types"
	"github.com/troian/surgemq/server"
	"github.com/troian/surgemq/tests/mqtt/config"
	"github.com/troian/surgemq/tests/mqtt/proxy"
	"github.com/troian/surgemq/tests/mqtt/test1"
	"github.com/troian/surgemq/tests/mqtt/test10"
	"github.com/troian/surgemq/tests/mqtt/test2"
	"github.com/troian/surgemq/tests/mqtt/test4"
	_ "github.com/troian/surgemq/topics/mem"
	"github.com/troian/surgemq/types"
	"os"
	"testing"
)

type internalAuth struct {
	creds map[string]string
}

type testEntry struct {
	name  string
	entry func(t *testing.T)
}

var testList []testEntry

func init() {
	testList = []testEntry{
		{
			name:  "single threaded client using receive",
			entry: test1.SubTest1,
		},
		{
			name:  "multi-threaded client using callbacks",
			entry: test1.SubTest2,
		},
		{
			name:  "connack return codes",
			entry: test1.SubTest3,
		},
		//{
		//	name:  "persistence",
		//	entry: test1.SubTest4,
		//},
		//{
		//	name:  "disconnect with quiesce timeout should allow exchanges to complete",
		//	entry: test1.SubTest5,
		//},
		//{
		//	name:  "connection lost and will messages",
		//	entry: test1.SubTest6,
		//},
		{
			name:  "multiple threads using same client object",
			entry: test2.SubTest1,
		},
		{
			name:  "multiple threads using callbacks",
			entry: test2.SubTest2,
		},
		{
			name:  "asynchronous connect",
			entry: test4.SubTest1,
		},
		{
			name:  "connect timeout",
			entry: test4.SubTest2,
		},
		{
			name:  "retained messages",
			entry: test10.SubTest1,
		},
		{
			name:  "offline queuing",
			entry: test10.SubTest2,
		},
		{
			name:  "overlapping test",
			entry: test10.SubTest3,
		},
	}
}

func (a internalAuth) Password(user, password string) error {
	if hash, ok := a.creds[user]; ok {
		if password == hash {
			return nil
		}
	}
	return auth.ErrAuthFailure
}

// nolint: golint
func (a internalAuth) AclCheck(clientID, user, topic string, access authTypes.AccessType) error {
	return auth.ErrAuthFailure
}

func (a internalAuth) PskKey(hint, identity string, key []byte, maxKeyLen int) error {
	return auth.ErrAuthFailure
}

func TestMain(m *testing.M) {
	config.Set(config.Provider{
		Host:         "tcp://localhost:1883",
		ProxyHost:    "tcp://localhost:1884",
		TestUser:     "testuser",
		TestPassword: "testpassword",
		MqttSasTopic: "MQTTSAS topic",
	})

	ia := internalAuth{
		creds: make(map[string]string),
	}

	ia.creds["testuser"] = "testpassword"

	var err error

	if err = auth.Register("internal", ia); err != nil {
		os.Exit(1)
	}

	var srv server.Type

	srv, err = server.New(server.Config{
		KeepAlive:      types.DefaultKeepAlive,
		AckTimeout:     types.DefaultAckTimeout,
		ConnectTimeout: 5,
		TimeoutRetries: types.DefaultTimeoutRetries,
		TopicsProvider: types.DefaultTopicsProvider,
		Authenticators: "internal",
		Anonymous:      true,
		Persistence: &persistType.BoltDBConfig{
			File: "./persist.db",
		},
		DupConfig: types.DuplicateConfig{
			Replace:   true,
			OnAttempt: nil,
		},
	})
	if err != nil {
		os.Exit(1)
	}

	var authMng *auth.Manager

	if authMng, err = auth.NewManager("internal"); err != nil {
		return
	}

	config := &server.Listener{
		Scheme:      "tcp4",
		Host:        "",
		Port:        1883,
		AuthManager: authMng,
	}

	if err = srv.ListenAndServe(config); err != nil {
		os.Exit(1)
	}

	prx, _ := proxy.NewProxy("localhost:1884", "localhost:1883")

	res := m.Run()

	prx.Shutdown()

	if err = srv.Close(); err != nil {

	}

	os.Exit(res)
}

func Test(t *testing.T) {
	for _, test := range testList {
		t.Run(test.name, test.entry)
	}
}
