package agent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/internal/engine"
	"github.com/camptocamp/bivac/internal/utils"
)

func Backup(targetURL, backupPath, hostname string, force bool, logReceiver string) {
	e := &engine.Engine{
		DefaultArgs: []string{
			"--no-cache",
			"--json",
			"-r",
			targetURL,
		},
		Output: make(map[string]utils.OutputFormat),
	}

	output := e.Backup(backupPath, hostname, force)

	if logReceiver != "" {
		data := `{"data":` + output + `}`
		req, err := http.NewRequest("POST", logReceiver, bytes.NewBuffer([]byte(data)))
		if err != nil {
			log.Errorf("failed to build new request: %s\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+os.Getenv("BIVAC_SERVER_PSK"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("failed to send data: %s\n", err)
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("failed to read body: %s\n", err)
			return
		}
		if resp.StatusCode != 200 {
			log.Infof("Response from API: %s", b)
		}
		return
	}
	fmt.Println(output)
	return
}

func Restore(targetURL, backupPath, hostname string) {
	return
}
