package downloadergo

import (
	"fmt"
	"github.com/gookit/slog"
	"github.com/ohanakogo/errhandlergo"
	"net/http"
	"os"
	"sync"
	"time"
)

type mThreadResult struct {
	Done bool
	Err  error
}

func (task *Task) baseDownload() {
	resp, err := http.Head(task.URL)
	if err != nil {
		task.progressMonitor.push(err, task.status)
		return
	}
	defer func() {
		errhandlergo.HandleError(resp.Body.Close(), nil)
	}()

	// create target file
	file, err := os.Create(task.FilePath)
	if err != nil {
		task.progressMonitor.push(err, task.status)
		return
	}
	errhandlergo.HandleError(file.Close(), nil)

	task.status.mContentLength = resp.ContentLength

	if task.Config.Debug {
		slog.Debug("Task Information:")
		slog.Debugf("Thread count: %d", task.Config.ThreadCount)
		slog.Debugf("Buffer size: %d bytes", task.Config.BufferSize)
		slog.Debugf("Timeout: %d ms", task.Config.Timeout/time.Millisecond)
		slog.Debugf("Max retries: %d", task.Config.MaxRetries)

		slog.Debugf("Content-Length: %d", task.status.mContentLength)
	}

	threadCount := task.Config.ThreadCount
	var wg sync.WaitGroup

	partSize := (task.status.mContentLength + int64(threadCount) - 1) / int64(threadCount)

	for i := 0; i < func() int {
		if task.status.mContentLength >= 0 {
			return threadCount
		}
		// if we can not get content length, try to download by single thread
		return 0
	}(); i++ {
		dRange := mPair{0, 0}

		// try to calculate start and end range
		if task.status.mContentLength >= 0 {
			dRange.start = int64(i) * partSize
			dRange.end = dRange.start + partSize
			// end byte overflowing, reset to last position of bytes
			if dRange.end > task.status.mContentLength {
				dRange.end = task.status.mContentLength
			}

			if dRange.start >= dRange.end {
				continue
			}
		}

		wg.Add(1)
		task.status.mWorkingRoutines++
		go func(routineNumber int) {
			defer func() {
				wg.Done()
				task.status.mWorkingRoutines--
				if task.Config.Debug {
					slog.Debugf("stopping routine %d", routineNumber)
				}
			}()
			if task.Config.Debug {
				slog.Debugf("creating routine %d", routineNumber)
			}
			retryCount := 0
			for {
				select {
				case <-time.After(time.Second):
					res := mThreadResult{}

					task.baseDownloadPart(dRange)
					res = <-task.status.mThreadResultChan
					if res.Err == nil {
						if res.Done {
							return
						}
					} else {
						slog.Errorf("thread %d: %s", routineNumber, res.Err.Error())
						if retryCount >= task.Config.MaxRetries {
							task.progressMonitor.push(fmt.Errorf("failed after %d retries", retryCount), task.status)
							return
						}
						retryCount++
					}
				}
			}
		}(task.status.mWorkingRoutines)
	}

	task.progressMonitor.sync()

	wg.Wait()
}
