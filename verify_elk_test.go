package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"fmt"
	"time"
)

func TestVerifyElkExpectedNoOfMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkExpectedNoOfMatches(string(output), 0)
	assert.Equal(5, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches(string(output), 1)
	assert.Equal(5, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches(string(output), 5)
	assert.Equal(0, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkExpectedNoOfMatchesWithNoMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk_no_matches.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkExpectedNoOfMatches(string(output), 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches(string(output), 1)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches(string(output), 5)
	assert.Equal(1, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkAtLeastNoOfMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkAtLeastNoOfMatches(string(output), 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 1)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 5)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 6)
	assert.Equal(2, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 55)
	assert.Equal(2, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkAtLeastNoOfMatchesWithNoMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk_no_matches.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkAtLeastNoOfMatches(string(output), 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 1)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 5)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 6)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches(string(output), 55)
	assert.Equal(1, len(errors), fmt.Sprint(errors))
}

func TestFormatDateForElkIndex(t *testing.T) {
	assert := assert.New(t)

	time, err := time.Parse("2006-01-02", "2010-10-10")
	assert.Nil(err, fmt.Sprint(err))

	assert.Equal("2010.10.10", formatDateForElkIndex(time))
}
