package downloadergo

// NewDownloaderTask create a task to download file
func NewDownloaderTask(fileUrl string, filePath string, config *TaskConfig) (task *Task) {
	task = &Task{
		URL:      fileUrl,
		FilePath: filePath,
		Config:   config,
	}
	task.prepare()
	return task
}

// Download run download task with goroutine
func (task *Task) Download() {
	task.waitGroup.Add(1)
	go func() {
		defer task.waitGroup.Done()
		task.baseDownload()
	}()
}

// Wait waiting for download finish
func (task *Task) Wait() {
	task.waitGroup.Wait()
}
