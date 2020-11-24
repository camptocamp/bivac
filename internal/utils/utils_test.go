package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDockerAgentImage(t *testing.T) {
	testCases := []struct {
		givenManagerVersion  string
		expectedAgentVersion string
	}{
		{
			"2.2.1-ad68ec-dirty",
			"latest",
		},
		{
			"2.2.1",
			"2.2.1",
		},
		{
			"",
			"latest",
		},
		{
			"2.1.0-rc0",
			"2.1.0-rc0",
		},
	}

	for _, testCase := range testCases {
		agentVersion := ComputeDockerAgentImage(testCase.givenManagerVersion)
		assert.Equal(t, agentVersion, testCase.expectedAgentVersion)
	}
}
