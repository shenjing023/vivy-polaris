package options

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shenjing023/vivy-polaris/errors"

	"golang.org/x/time/rate"
)

const (
	CUSTOM_ERROR1 errors.Code = 100
	CUSTOM_ERROR2 errors.Code = 101
)

func TestServiceConfig(t *testing.T) {
	// errors.AppendErrCode("CUSTOM_ERROR1", CUSTOM_ERROR1)
	// errors.AppendErrCode("CUSTOM_ERROR2", CUSTOM_ERROR2)
	retry := RetryPolicy{
		MaxAttempts:          3,
		MaxBackoff:           "1s",
		InitialBackoff:       "1s",
		BackoffMultiplier:    5,
		RetryableStatusCodes: []string{"CUSTOM_ERROR1", "CUSTOM_ERROR2"},
	}
	mc := MethodConfig{
		Name: []MethodName{
			{"service1", "method1"},
			{"service2", "method2"},
		},
		RetryPolicy: retry,
	}
	sc := ServiceConfig{
		Methodconfig:        []MethodConfig{mc},
		LoadBalancingPolicy: "round_robin",
	}
	r, err := json.Marshal(&sc)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(r))
	sc2 := ServiceConfig{}
	r, err = json.Marshal(&sc2)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(r))
}

func TestRateLimit(t *testing.T) {
	limiter := rate.NewLimiter(10, 10)
	for i := 0; i < 20; i++ {
		t.Log(limiter.Allow())
	}
	time.Sleep(time.Second * 1)
	for i := 0; i < 20; i++ {
		t.Log(limiter.Allow())
	}
}
