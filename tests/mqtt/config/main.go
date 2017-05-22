package config

import "testing"

type Provider struct {
	Host         string
	ProxyHost    string
	TestUser     string
	TestPassword string
	MqttSasTopic string
}

var p Provider

func Set(c Provider) {
	p = c
}

func Get() Provider {
	return p
}

type TestingWrap interface {
	Name() string
	Configure()
	Run(t *testing.T)
}
