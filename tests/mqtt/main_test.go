package tests

import (
	"os"
	"testing"

	"github.com/troian/surgemq/auth"
	authTypes "github.com/troian/surgemq/auth/types"
	persistType "github.com/troian/surgemq/persistence/types"
	"github.com/troian/surgemq/server"
	"github.com/troian/surgemq/tests/mqtt/config"
	"github.com/troian/surgemq/tests/mqtt/proxy"

	"fmt"
	testTypes "github.com/troian/surgemq/tests/types"
	_ "github.com/troian/surgemq/topics/mem"
	"github.com/troian/surgemq/types"
	"time"
)

type internalAuth struct {
	creds map[string]string
}

var testList []testTypes.Provider

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
		fmt.Println(err.Error())
		os.Exit(1)
	}

	<-time.After(1 * time.Second)
	prx, _ := proxy.NewProxy("localhost:1884", "localhost:1883")

	res := m.Run()

	prx.Shutdown()

	if err = srv.Close(); err != nil {

	}

	os.Exit(res)
}

func Test(t *testing.T) {
	for _, entry := range testList {
		t.Run(entry.Name(), entry.Run)
	}
}
