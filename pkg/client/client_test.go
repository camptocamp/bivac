package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/camptocamp/bivac/pkg/volume"
)

// NewClient
func TestNewClientValid(t *testing.T) {
	// Prepare test
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedClient := &Client{
		remoteAddress: "http://fakeserver",
		psk:           "psk",
	}

	// Run test
	httpmock.RegisterResponder("GET", "http://fakeserver/ping",
		httpmock.NewStringResponder(200, `{"type": "pong"}`))

	c, err := NewClient("http://fakeserver", "psk")

	assert.Nil(t, err)
	assert.Equal(t, c, expectedClient)
}

func TestNewClientFailedToConnect(t *testing.T) {
	// Prepare test
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedError := errors.New("failed to connect")
	expectedClient := &Client{
		remoteAddress: "http://fakefakeserver",
		psk:           "psk",
	}

	// Run test cases
	c, err := NewClient("http://fakefakeserver", "psk")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Equal(t, c, expectedClient)
}

func TestNewClientWrongResponse(t *testing.T) {
	// Prepare test
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedClient := &Client{
		remoteAddress: "http://fakeserver",
		psk:           "psk",
	}

	// Run test
	httpmock.RegisterResponder("GET", "http://fakeserver/ping",
		httpmock.NewStringResponder(200, `{"type": "foo"}`))

	c, err := NewClient("http://fakeserver", "psk")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong response")
	assert.Equal(t, c, expectedClient)
}

func TestNewClientFailedToUnmarshal(t *testing.T) {
	// Prepare test
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedClient := &Client{
		remoteAddress: "http://fakeserver",
		psk:           "psk",
	}

	// Run test
	httpmock.RegisterResponder("GET", "http://fakeserver/ping",
		httpmock.NewStringResponder(200, ``))

	c, err := NewClient("http://fakeserver", "psk")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
	assert.Equal(t, c, expectedClient)
}

func TestNewClientWrongStatusCode(t *testing.T) {
	// Prepare test
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	expectedClient := &Client{
		remoteAddress: "http://fakeserver",
		psk:           "psk",
	}

	// Run test
	httpmock.RegisterResponder("GET", "http://fakeserver/ping",
		httpmock.NewStringResponder(404, ``))

	c, err := NewClient("http://fakeserver", "psk")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong status code")
	assert.Equal(t, c, expectedClient)
}

// GetVolumes
func TestGetVolumesValid(t *testing.T) {
	// Prepare test
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	fakeResponse := `[
		{
			"id": "foo",
			"name": "foo",
			"mountpoint": "/foo"
		},
		{
			"id": "bar",
			"name": "bar",
			"mountpoint": "/bar"
		}
	]`

	expectedVolumes := []volume.Volume{
		volume.Volume{
			ID:         "foo",
			Name:       "foo",
			Mountpoint: "/foo",
		},
		volume.Volume{
			ID:         "bar",
			Name:       "bar",
			Mountpoint: "/bar",
		},
	}

	// Run test
	httpmock.RegisterResponder("GET", "http://fakeserver/volumes",
		httpmock.NewStringResponder(200, fakeResponse))

	c := &Client{
		remoteAddress: "http://fakeserver",
		psk:           "psk",
	}
	volumes, err := c.GetVolumes()

	assert.Nil(t, err)
	assert.Equal(t, volumes, expectedVolumes)
}
