package downloadergo

import (
	"fmt"
	"testing"
	"time"
)

func speedMonitor(ch chan TaskProgress) {
	for {
		select {
		case res := <-ch:
			if res.Err == nil {
				fmt.Printf("\rDownloading... %.2f%%, working threads: %d, speed: %.2f MB/s", res.Progress, res.WorkingThreads, res.SpeedPerSecond/float64(1024*1024))
				if res.Done {
					fmt.Print("\n")
					return
				}
			} else {
				fmt.Println(res.Err)
				return
			}
		}
	}
}

func TestTask_Download(t *testing.T) {
	task := NewDownloaderTask("https://pkg.biligame.com/games/blhx_6.2.1_bilibili_20221107_143634.apk", "test.apk", &TaskConfig{
		Debug:       true,
		MaxRetries:  3,
		Proxy:       nil,
		BufferSize:  1024 * 8,
		ThreadCount: 16,
	})
	ch := task.EnableProgressMonitor(time.Second / 4)
	task.Download()

	go speedMonitor(ch)

	task.Wait()
}
