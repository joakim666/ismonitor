package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"fmt"
)

func TestVerifyRunningDockerContainers(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_docker.txt")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyRunningDockerContainers(string(output), []string {"confluence", "cassandra", "postgres"})
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyRunningDockerContainers(string(output), []string {"confluence", "cassandra", "postgres", "foo"})
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyRunningDockerContainers(string(output), []string {"confluence", "cassandra", "postgres", "foo", "bar"})
	assert.Equal(2, len(errors), fmt.Sprint(errors))

	errors = verifyRunningDockerContainers(string(output), []string {"foo", "bar"})
	assert.Equal(2, len(errors), fmt.Sprint(errors))
}
