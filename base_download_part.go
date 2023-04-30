package downloadergo

import (
	"fmt"
	"github.com/gookit/slog"
	"github.com/ohanakogo/errhandlergo"
	"github.com/ohanakogo/ohanakoutilgo"
	"io"
	"net/http"
	"os"
)

type mPair struct {
	start int64
	end   int64
}

func (task *Task) baseDownloadPart(dRange mPair) {
	defer func() {
		errhandlergo.HandleRecover(func(err any) {
			ohanakoutilgo.CastThen[error](err, func(err error) {
				task.status.mThreadResultChan <- mThreadResult{Err: err}
				slog.Errorf("download task failed")
			})
		})
	}()

	req, err := http.NewRequest("GET", task.URL, nil)
	errhandlergo.HandleError(err, func() { panic(err) })

	if dRange.start < dRange.end {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", dRange.start, dRange.end-1))
	}

	client := task.getClient()
	resp, err := client.Do(req)
	errhandlergo.HandleError(err, func() { panic(err) })

	defer func() {
		errhandlergo.HandleError(resp.Body.Close(), nil)
	}()

	file, err := os.OpenFile(task.FilePath, os.O_WRONLY, 0644)
	errhandlergo.HandleError(err, func() { panic(err) })
	defer func() {
		errhandlergo.HandleError(file.Close(), nil)
	}()

	buffer := task.getBuffer()
	offset := dRange.start
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			written, err := file.WriteAt(buffer[:n], offset)
			errhandlergo.HandleError(err, func() { panic(err) })
			offset += int64(written)
			task.status.mDownloadedSize += int64(written)
		}
		if err == io.EOF {
			task.status.mThreadResultChan <- mThreadResult{Done: true}
			return
		}
		if err != nil {
			task.status.mThreadResultChan <- mThreadResult{Err: err}
			task.status.mDownloadedSize -= offset - dRange.start
			task.progressMonitor.fallback(offset - dRange.start)
			return
		}
	}
}
