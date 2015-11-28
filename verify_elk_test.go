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

	errors := verifyElkExpectedNoOfMatches([]string{string(output)}, 0)
	assert.Equal(5, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches([]string{string(output)}, 1)
	assert.Equal(5, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches([]string{string(output)}, 5)
	assert.Equal(0, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkExpectedNoOfMatchesWithNoMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk_no_matches.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkExpectedNoOfMatches([]string{string(output)}, 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches([]string{string(output)}, 1)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches([]string{string(output)}, 5)
	assert.Equal(1, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkAtLeastNoOfMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkAtLeastNoOfMatches([]string{string(output)}, 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 1)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 5)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 6)
	assert.Equal(2, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 55)
	assert.Equal(2, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkAtLeastNoOfMatchesWithNoMatches(t *testing.T) {
	assert := assert.New(t)

	output, err := ioutil.ReadFile("test/output_elk_no_matches.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkAtLeastNoOfMatches([]string{string(output)}, 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 1)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 5)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 6)
	assert.Equal(1, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output)}, 55)
	assert.Equal(1, len(errors), fmt.Sprint(errors))
}

func TestVerifyElkAroundMidnight(t *testing.T) {
	assert := assert.New(t)

	output1, err := ioutil.ReadFile("test/output_elk_before_midnight.json")
	assert.Nil(err, fmt.Sprint(err))

	output2, err := ioutil.ReadFile("test/output_elk_after_midnight.json")
	assert.Nil(err, fmt.Sprint(err))

	errors := verifyElkExpectedNoOfMatches([]string{string(output1), string(output2)}, 0)
	assert.Equal(5, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches([]string{string(output1), string(output2)}, 1)
	assert.Equal(5, len(errors), fmt.Sprint(errors))

	errors = verifyElkExpectedNoOfMatches([]string{string(output1), string(output2)}, 5)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output1), string(output2)}, 0)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output1), string(output2)}, 1)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output1), string(output2)}, 5)
	assert.Equal(0, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output1), string(output2)}, 6)
	assert.Equal(2, len(errors), fmt.Sprint(errors))

	errors = verifyElkAtLeastNoOfMatches([]string{string(output1), string(output2)}, 55)
	assert.Equal(2, len(errors), fmt.Sprint(errors))

}

func TestFormatDateForElkIndex(t *testing.T) {
	assert := assert.New(t)

	time, err := time.Parse("2006-01-02", "2010-10-10")
	assert.Nil(err, fmt.Sprint(err))

	assert.Equal("2010.10.10", formatDateForElkIndex(time))
}

func TestElkIndexToUse(t *testing.T) {
	assert := assert.New(t)

	time1, err := time.Parse("2006-01-02 15:04:05", "2010-10-10 11:11:12")
	assert.Nil(err, fmt.Sprint(err))
	indexes := elkIndexToUse(time1, 5)
	assert.Equal(1, len(indexes))
	assert.Equal("2010.10.10", indexes[0])

	time2, err := time.Parse("2006-01-02 15:04:05", "2010-10-10 00:01:12")
	assert.Nil(err, fmt.Sprint(err))
	indexes = elkIndexToUse(time2, 5)
	assert.Equal(2, len(indexes))
	assert.Equal("2010.10.09", indexes[0])
	assert.Equal("2010.10.10", indexes[1])
}
