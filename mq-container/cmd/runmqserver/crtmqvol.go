/*
© Copyright IBM Corporation 2017, 2019

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
package main

import (
	"os"
	"runtime"
	"syscall"

	"github.com/ibm-messaging/mq-container/internal/command"
)

func createVolume(dataPath string) error {
	fi, err := os.Stat(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// #nosec G301
			err = os.MkdirAll(dataPath, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	fi, err = os.Stat(dataPath)
	if err != nil {
		return err
	}
	sys := fi.Sys()
	if sys != nil && runtime.GOOS == "linux" {
		stat := sys.(*syscall.Stat_t)
		mqmUID, mqmGID, err := command.LookupMQM()
		if err != nil {
			return err
		}
		log.Debugf("mqm user is %v (%v)", mqmUID, mqmGID)
		if int(stat.Uid) != mqmUID || int(stat.Gid) != mqmGID {
			err = os.Chown(dataPath, mqmUID, mqmGID)
			if err != nil {
				log.Printf("Error: Unable to change ownership of %v", dataPath)
				return err
			}
		}
	}
	return nil
}
