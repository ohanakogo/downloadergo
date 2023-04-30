package downloadergo

import (
	"github.com/gookit/slog"
	"time"
)

type TaskProgress struct {
	SpeedPerSecond float64
	ContentSize    int64
	Done           bool
	DownloadedSize int64
	WorkingThreads int
	Progress       float64
	Err            error
}

type mProgressMonitor struct {
	syncRoutine func()

	ch chan TaskProgress

	lastBytes      int64
	currentBytes   int64
	speedPerSecond float64
}

func (task *Task) EnableProgressMonitor(refreshInterval time.Duration) chan TaskProgress {
	task.progressMonitor = &mProgressMonitor{
		syncRoutine: func() {
			ticker := time.NewTicker(refreshInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					// calculate speedPerSecond
					task.progressMonitor.currentBytes = task.status.mDownloadedSize
					task.progressMonitor.speedPerSecond = float64(task.progressMonitor.currentBytes-task.progressMonitor.lastBytes) / (float64(refreshInterval) / float64(time.Second))
					task.progressMonitor.lastBytes = task.progressMonitor.currentBytes

					task.progressMonitor.push(nil, task.status)
					if task.status.mWorkingRoutines == 0 {
						return
					}
				}
			}
		},
		ch: make(chan TaskProgress),
	}
	return task.progressMonitor.ch
}

func (p *mProgressMonitor) push(err error, status *TaskStatus) {
	if p != nil {
		select {
		case p.ch <- TaskProgress{
			SpeedPerSecond: p.speedPerSecond,
			ContentSize:    status.mContentLength,
			Done:           status.mWorkingRoutines == 0,
			DownloadedSize: status.mDownloadedSize,
			WorkingThreads: status.mWorkingRoutines,
			Progress: func() float64 {
				if status.mContentLength > 0 {
					return float64(status.mDownloadedSize) * 100 / float64(status.mContentLength)
				}
				return 0
			}(),
			Err: err,
		}:
			break
		default:
			slog.Warnf("downloadergo: progress sync channel is full")
		}

	}
}

func (p *mProgressMonitor) fallback(fallbackBytes int64) {
	p.lastBytes -= fallbackBytes
	p.currentBytes -= fallbackBytes
}

func (p *mProgressMonitor) sync() {
	if p != nil && p.syncRoutine != nil {
		go p.syncRoutine()
	}
}
