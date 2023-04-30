package downloadergo

import (
	"net/http"
	"net/url"
	"time"
)

func (task *Task) prepare() {
	checkThreadCount := func() {
		if task.Config.ThreadCount <= 0 {
			task.Config.ThreadCount = 1
		}
	}
	checkTimeout := func() {
		if task.Config.Timeout <= 0 {
			task.Config.Timeout = time.Millisecond * 5000
		}
	}
	initStatus := func() {
		task.status = &TaskStatus{
			mThreadResultChan: make(chan mThreadResult, task.Config.ThreadCount+1),
		}
	}

	checkThreadCount()
	checkTimeout()
	initStatus()
}

func (task *Task) getBuffer() []byte {
	bufferSize := func() int {
		if task.Config.BufferSize != 0 {
			return task.Config.BufferSize
		}
		// Default is 4096 bytes per part
		return 1024 * 4
	}()
	return make([]byte, bufferSize)
}

func (task *Task) getClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: func() func(r *http.Request) (*url.URL, error) {
				if task.Config.Proxy == nil {
					return nil
				}
				return http.ProxyURL(task.Config.Proxy)
			}(),
		},
	}
}
