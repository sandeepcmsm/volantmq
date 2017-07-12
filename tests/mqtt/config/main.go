package config

import "testing"

// nolint: golint
type Provider struct {
	Host         string
	ProxyHost    string
	TestUser     string
	TestPassword string
	MqttSasTopic string
}

var p Provider

// nolint: golint
func Set(c Provider) {
	p = c
}

// nolint: golint
func Get() Provider {
	return p
}

// nolint: golint
type TestingWrap interface {
	Name() string
	Configure()
	Run(t *testing.T)
}
