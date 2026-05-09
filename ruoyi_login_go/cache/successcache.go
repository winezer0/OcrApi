package cache

import (
	"fmt"
	"ruoyi_login_go/filewriter"
	"ruoyi_login_go/logging"
	"sync"
)

// SuccessCache 成功记录缓存（类似 FailedCache 的设计，避免全局变量竞态）
type SuccessCache struct {
	writer   *filewriter.FileWriter
	writerMu sync.Mutex
}

// NewSuccessCache 创建成功记录缓存
func NewSuccessCache() *SuccessCache {
	return &SuccessCache{}
}

// Save 保存成功记录到缓存文件
func (sc *SuccessCache) Save(successFile, username, password string) error {
	writer, err := sc.getWriter(successFile)
	if err != nil {
		return err
	}

	entry := fmt.Sprintf("%s:%s\n", username, password)
	if err := writer.Write(entry); err != nil {
		logging.Warnf("写入成功记录失败: %v", err)
	}
	return nil
}

// getWriter 获取或创建异步写入器
func (sc *SuccessCache) getWriter(successFile string) (*filewriter.FileWriter, error) {
	sc.writerMu.Lock()
	defer sc.writerMu.Unlock()

	if sc.writer == nil {
		var err error
		sc.writer, err = filewriter.NewFileWriter(successFile)
		if err != nil {
			return nil, fmt.Errorf("创建成功记录异步写入器失败: %w", err)
		}
	}

	return sc.writer, nil
}

// Close 关闭异步写入器（程序退出时调用）
func (sc *SuccessCache) Close() error {
	sc.writerMu.Lock()
	defer sc.writerMu.Unlock()

	if sc.writer != nil {
		err := sc.writer.Close()
		sc.writer = nil
		return err
	}
	return nil
}
