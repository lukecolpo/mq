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
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// waitForFile waits until the specified file exists
func waitForFile(ctx context.Context, path string) (os.FileInfo, error) {
	var fi os.FileInfo
	var err error
	// Wait for file to exist
	for {
		select {
		// Check to see if cancellation has been requested
		case <-ctx.Done():
			return os.Stat(path)
		default:
			fi, err = os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					time.Sleep(500 * time.Millisecond)
					continue
				} else {
					return nil, fmt.Errorf("mirror: unable to get info on file %v", path)
				}
			}
			return fi, nil
		}
	}
}

type mirrorFunc func(msg string, isQMLog bool) bool

// mirrorAvailableMessages prints lines from the file, until no more are available
func mirrorAvailableMessages(f *os.File, mf mirrorFunc, isQMLog bool) {
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		t := scanner.Text()
		if mf(t, isQMLog) {
			count++
		}
	}
	if count > 0 {
		log.Debugf("Mirrored %v log entries from %v", count, f.Name())
	}
	err := scanner.Err()
	if err != nil {
		log.Errorf("Error reading file %v: %v", f.Name(), err)
		return
	}
}

// mirrorLog tails the specified file, and logs each line to stdout.
// This is useful for usability, as the container console log can show
// messages from the MQ error logs.
func mirrorLog(ctx context.Context, wg *sync.WaitGroup, path string, fromStart bool, mf mirrorFunc, isQMLog bool) (chan error, error) {
	errorChannel := make(chan error, 1)
	var offset int64 = -1
	var f *os.File
	var err error
	var fi os.FileInfo
	// Need to check if the file exists before returning, otherwise we have a
	// race to see if the new file get created before we can test for it
	fi, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, so ensure we start at the beginning
			offset = 0
		} else {
			return nil, err
		}
	} else {
		// If the file exists, open it now, before we return.  This makes sure
		// the file is open before the queue manager is created or started.
		// Otherwise, there would be the potential for a nearly-full file to
		// rotate before the goroutine had a chance to open it.
		f, err = os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return nil, err
		}
		// File already exists, so start reading at the end
		offset = fi.Size()
	}
	// Increment wait group counter, only if the goroutine gets started
	wg.Add(1)
	go func() {
		// Notify the wait group when this goroutine ends
		defer func() {
			log.Debugf("Finished monitoring %v", path)
			wg.Done()
		}()
		if f == nil {
			// File didn't exist, so need to wait for it
			fi, err = waitForFile(ctx, path)
			if err != nil {
				log.Error(err)
				errorChannel <- err
				return
			}
			if fi == nil {
				return
			}
			log.Debugf("File exists: %v, %v", path, fi.Size())
			f, err = os.OpenFile(path, os.O_RDONLY, 0)
			if err != nil {
				log.Error(err)
				errorChannel <- err
				return
			}
		}

		fi, err = f.Stat()
		if err != nil {
			log.Error(err)
			errorChannel <- err
			return
		}
		// The file now exists.  If it didn't exist before we started, offset=0
		// Always start at the beginning if we've been told to go from the start
		if offset != 0 && !fromStart {
			log.Debugf("Seeking offset %v in file %v", offset, path)
			_, err = f.Seek(offset, 0)
			if err != nil {
				log.Errorf("Unable to return to offset %v: %v", offset, err)
			}
		}
		closing := false
		for {
			// If there's already data there, mirror it now.
			mirrorAvailableMessages(f, mf, isQMLog)
			// Wait for the new log file (after rotation)
			newFI, err := waitForFile(ctx, path)
			if err != nil {
				log.Error(err)
				errorChannel <- err
				return
			}
			if !os.SameFile(fi, newFI) {
				log.Debugf("Detected log rotation in file %v", path)
				// WARNING: There is a possible race condition here.  If *another*
				// log rotation happens before we can open the new file, then we
				// could skip all those messages.  This could happen with a very small
				// MQ error log size.
				mirrorAvailableMessages(f, mf, isQMLog)
				err = f.Close()
				if err != nil {
					log.Errorf("Unable to close mirror file handle: %v", err)
				}
				// Re-open file
				log.Debugf("Re-opening error log file %v", path)
				f, err = os.OpenFile(path, os.O_RDONLY, 0)
				if err != nil {
					log.Error(err)
					errorChannel <- err
					return
				}
				fi = newFI
				// Don't seek this time, because we know it's a new file
				mirrorAvailableMessages(f, mf, isQMLog)
			}
			select {
			case <-ctx.Done():
				log.Debugf("Context cancelled for mirroring %v", path)
				if closing {
					log.Debugf("Shutting down mirror for %v", path)
					return
				}
				// Set a flag, to allow one more time through the loop
				closing = true
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	return errorChannel, nil
}
