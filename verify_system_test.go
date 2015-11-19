package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"fmt"
)

func TestVerifyFreeSpace(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_df.txt")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyFreeSpace(string(output), 80)
	assert.Equal(0, len(errors), "Should not be any mount with more than 80% usage")

	errors = verifyFreeSpace(string(output), 40)
	assert.Equal(1, len(errors), "Should be one mount with more than 40% usage")

	errors = verifyFreeSpace(string(output), 10)
	assert.Equal(2, len(errors), "Should be two mounts with more than 40% usage")
}

func TestVerifyUptimeLoad(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_uptime.txt")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyUptimeLoad(string(output), 5)
	assert.Equal(0, len(errors), "The load isn't over 5")

	errors = verifyUptimeLoad(string(output), 0)
	assert.Equal(1, len(errors), "The load is over 0")

	errors = verifyUptimeLoad(string(output), 0.1)
	assert.Equal(1, len(errors), "The load is over 0")
}

