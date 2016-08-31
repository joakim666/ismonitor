package main

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyFreeSpace(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_df.txt")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyFreeSpace(string(output), 80)
	assert.Equal(0, len(errors), "Should not be any mount with more than 80% usage")

	errors = verifyFreeSpace(string(output), 40)
	assert.Equal(1, len(errors), "Should be one mount with more than 40% usage")
	assert.Equal("Disk usage verification error", errors[0].title)
	assert.Equal("Disk usage of / at 42 percent\n", errors[0].message)

	errors = verifyFreeSpace(string(output), 10)
	assert.Equal(2, len(errors), "Should be two mounts with more than 10% usage")
	assert.Equal("Disk usage verification error", errors[0].title)
	assert.Equal("Disk usage of / at 42 percent\n", errors[0].message)
	assert.Equal("Disk usage verification error", errors[1].title)
	assert.Equal("Disk usage of /boot at 15 percent\n", errors[1].message)
}

func TestVerifyLoadAvg(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_proc_loadavg.txt")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyLoadAvg(string(output), 5)
	assert.Equal(0, len(errors), "The load isn't over 5")

	errors = verifyLoadAvg(string(output), 0)
	assert.Equal(1, len(errors), "The load is over 0")

	errors = verifyLoadAvg(string(output), 0.1)
	assert.Equal(1, len(errors), "The load is over 0")
}
