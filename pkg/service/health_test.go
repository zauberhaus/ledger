package service_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Health(t *testing.T) {
	resp, err := http.Get(url + "/health")
	if !assert.NoError(t, err) {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "ok", string(data))
}

func Test_Metrics(t *testing.T) {
	resp, err := http.Get(metricsUrl + "/metrics")
	if !assert.NoError(t, err) {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, strings.HasPrefix(string(data), "# HELP go_gc_duration_seconds"))
}
