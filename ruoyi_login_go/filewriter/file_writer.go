package filewriter

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"ruoyi_login_go/logging"
)

// 写入超时时间
const writeTimeout = 30 * time.Second

// FileWriter 异步文本行写入器
type FileWriter struct {
	file      *os.File
	bufWriter *bufio.Writer
	ch        chan string
	done      chan struct{}
	closeOnce sync.Once
	closed    bool
}

// NewFileWriter 创建异步文本行写入器
func NewFileWriter(filePath string) (*FileWriter, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	fw := &FileWriter{
		file:      f,
		bufWriter: bufio.NewWriter(f),
		ch:        make(chan string, 1000),
		done:      make(chan struct{}),
	}

	go fw.writeLoop()

	return fw, nil
}

// writeLoop 异步写入循环（单 goroutine 消费 channel，天然串行无需互斥锁）
func (fw *FileWriter) writeLoop() {
	for line := range fw.ch {
		if _, err := fw.bufWriter.WriteString(line); err != nil {
			logging.Warnf("文件写入失败: %v", err)
		}
	}
	fw.bufWriter.Flush()
	close(fw.done)
}

// Write 写入一行文本（队列满时阻塞等待，超时后返回错误；写入器已关闭时返回错误）
func (fw *FileWriter) Write(line string) error {
	if fw.closed {
		return fmt.Errorf("写入器已关闭")
	}

	select {
	case fw.ch <- line:
		return nil
	case <-time.After(writeTimeout):
		return fmt.Errorf("写入超时（%v 后仍未写入）", writeTimeout)
	}
}

// Close 关闭写入器并刷新缓冲区（使用 sync.Once 防止重复关闭 channel 和文件导致 panic）
func (fw *FileWriter) Close() error {
	var closeErr error
	fw.closeOnce.Do(func() {
		fw.closed = true
		close(fw.ch)
		<-fw.done
		closeErr = fw.file.Close()
	})
	return closeErr
}
