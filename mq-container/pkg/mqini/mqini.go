/*
© Copyright IBM Corporation 2018, 2019

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mqini provides information about queue managers
package mqini

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ibm-messaging/mq-container/internal/command"
)

// QueueManager describe high-level configuration information for a queue manager
type QueueManager struct {
	Name             string
	Prefix           string
	Directory        string
	DataPath         string
	InstallationName string
}

// getQueueManagerFromStanza parses a queue manager stanza
func getQueueManagerFromStanza(stanza string) (*QueueManager, error) {
	scanner := bufio.NewScanner(strings.NewReader(stanza))
	qm := QueueManager{}
	for scanner.Scan() {
		l := scanner.Text()
		l = strings.TrimSpace(l)
		t := strings.Split(l, "=")
		switch t[0] {
		case "Name":
			qm.Name = t[1]
		case "Prefix":
			qm.Prefix = t[1]
		case "Directory":
			qm.Directory = t[1]
		case "DataPath":
			qm.DataPath = t[1]
		case "InstallationName":
			qm.InstallationName = t[1]
		}
	}
	return &qm, scanner.Err()
}

// GetQueueManager returns queue manager configuration information
func GetQueueManager(name string) (*QueueManager, error) {
	_, err := os.Stat("/var/mqm/mqs.ini")
	if err != nil {
		// Don't run dspmqinf, which will generate an FDC if mqs.ini isn't there yet
		return nil, errors.New("dspmqinf should not be run before crtmqdir")
	}
	// dspmqinf essentially returns a subset of mqs.ini, but it's simpler to parse
	out, _, err := command.Run("dspmqinf", "-o", "stanza", name)
	if err != nil {
		return nil, err
	}
	return getQueueManagerFromStanza(out)
}

// GetErrorLogDirectory returns the directory holding the error logs for the
// specified queue manager
func GetErrorLogDirectory(qm *QueueManager) string {
	return filepath.Join(GetDataDirectory(qm), "errors")
}

// GetDataDirectory returns the data directory for the specified queue manager
func GetDataDirectory(qm *QueueManager) string {
	if qm.DataPath != "" {
		// Data path has been set explicitly (e.g. for multi-instance queue manager)
		return qm.DataPath
	} else {
		return filepath.Join(qm.Prefix, "qmgrs", qm.Directory)
	}
}
