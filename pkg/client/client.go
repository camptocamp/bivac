package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/camptocamp/bivac/pkg/volume"
)

type Client struct {
	remoteAddress string
	psk           string
}

func NewClient(remoteAddress string, psk string) (c *Client, err error) {
	c = &Client{
		remoteAddress: remoteAddress,
		psk:           psk,
	}

	var pingResponse map[string]string
	err = c.newRequest(&pingResponse, "GET", "/ping", "")
	if err != nil {
		err = fmt.Errorf("failed to connect to the remote Bivac instance: %s", err)
		return
	}
	if pingResponse["type"] != "pong" {
		err = fmt.Errorf("wrong response from the Bivac instance: %v", pingResponse)
		return
	}
	return
}

func (c *Client) GetVolumes() (volumes []volume.Volume, err error) {
	err = c.newRequest(&volumes, "GET", "/volumes", "")
	if err != nil {
		err = fmt.Errorf("failed to connect to the remote Bivac instance: %s", err)
		return
	}
	return
}

func (c *Client) BackupVolume(volumeName string, force bool) (err error) {
	err = c.newRequest(nil, "POST", fmt.Sprintf("/backup/%s?force=%s", volumeName, strconv.FormatBool(force)), "")
	if err != nil {
		err = fmt.Errorf("failed to connect to the remote Bivac instance: %s", err)
		return
	}
	return
}

func (c *Client) RunRawCommand(volumeID string, cmd []string) (output string, err error) {
	var response map[string]interface{}

	postValue := make(map[string][]string)
	postValue["cmd"] = cmd

	postValueEncoded, _ := json.Marshal(postValue)

	err = c.newRequest(&response, "POST", fmt.Sprintf("/restic/%s", volumeID), string(postValueEncoded))
	if err != nil {
		err = fmt.Errorf("failed to connect to the remote Bivac instance: %s", err)
		return
	}

	output = response["data"].(string)
	return
}

func (c *Client) newRequest(data interface{}, method, endpoint, value string) (err error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, c.remoteAddress+endpoint, bytes.NewBuffer([]byte(value)))
	if err != nil {
		err = fmt.Errorf("failed to build request: %s", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.psk))

	res, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to send request: %s", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("failed to read body: %s", err)
		return
	}

	if res.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &data); err != nil {
			err = fmt.Errorf("failed to unmarshal response from the Bivac instance: %s", err)
			return err
		}
	} else {
		err = fmt.Errorf("received wrong status code from the Bivac instance: [%d] %s", res.StatusCode, string(body))
		return
	}
	return
}
