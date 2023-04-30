package downloadergo

import (
	"net/url"
	"sync"
	"time"
)

type Task struct {
	URL      string
	FilePath string
	Config   *TaskConfig

	status          *TaskStatus
	progressMonitor *mProgressMonitor
	waitGroup       sync.WaitGroup
}

type TaskConfig struct {
	Debug bool

	BufferSize  int
	MaxRetries  int
	Proxy       *url.URL
	ThreadCount int

	// Timeout timing when always reading 0 bytes from remote, cancel request if timeout
	Timeout time.Duration
}

type TaskStatus struct {
	mThreadResultChan chan mThreadResult

	mContentLength   int64
	mDownloadedSize  int64
	mWorkingRoutines int
}
