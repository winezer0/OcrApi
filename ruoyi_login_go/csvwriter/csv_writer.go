package csvwriter

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

	"ruoyi_login_go/logging"
)

// 写入超时时间
const writeTimeout = 30 * time.Second

// CSVWriter 异步 CSV 写入器
type CSVWriter struct {
	file      *os.File
	bufWriter *bufio.Writer
	writer    *csv.Writer
	ch        chan []string
	done      chan struct{}
	headers   []string
	closeOnce sync.Once
	closed    bool
}

// NewCSVWriter 创建异步 CSV 写入器
func NewCSVWriter(filePath string, headers []string) (*CSVWriter, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开 CSV 文件失败: %w", err)
	}

	bufWriter := bufio.NewWriter(f)

	w := &CSVWriter{
		file:      f,
		bufWriter: bufWriter,
		writer:    csv.NewWriter(bufWriter),
		ch:        make(chan []string, 1000),
		done:      make(chan struct{}),
		headers:   headers,
	}

	// 检查文件是否为空，空文件需要写入表头
	fileInfo, err := f.Stat()
	if err == nil && fileInfo.Size() == 0 {
		w.writer.Write(headers)
		w.writer.Flush()
		w.bufWriter.Flush()
	}

	// 启动异步写入 goroutine
	go w.writeLoop()

	return w, nil
}

// writeLoop 异步写入循环（单 goroutine 消费 channel，天然串行无需互斥锁）
func (w *CSVWriter) writeLoop() {
	for record := range w.ch {
		if err := w.writer.Write(record); err != nil {
			logging.Warnf("CSV 写入失败: %v", err)
		}
	}
	w.writer.Flush()
	w.bufWriter.Flush()
	close(w.done)
}

// Write 写入一行数据（队列满时阻塞等待，超时后返回错误）
func (w *CSVWriter) Write(record []string) error {
	if w.closed {
		return fmt.Errorf("CSV 写入器已关闭")
	}
	select {
	case w.ch <- record:
		return nil
	case <-time.After(writeTimeout):
		return fmt.Errorf("CSV 写入超时（%v 后仍未写入）", writeTimeout)
	}
}

// Close 关闭写入器并刷新缓冲区（使用 sync.Once 防止重复关闭 channel 导致 panic）
func (w *CSVWriter) Close() error {
	var closeErr error
	w.closeOnce.Do(func() {
		w.closed = true
		close(w.ch)
		<-w.done
		closeErr = w.file.Close()
	})
	return closeErr
}
