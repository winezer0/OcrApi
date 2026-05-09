package cache

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"ruoyi_login_go/filewriter"
	"ruoyi_login_go/logging"
)

// FailedCache 失败记录缓存
type FailedCache struct {
	entries  map[string]struct{}
	mu       sync.RWMutex
	writer   *filewriter.FileWriter
	writerMu sync.Mutex
}

// NewFailedCache 创建失败记录缓存
func NewFailedCache() *FailedCache {
	return &FailedCache{
		entries: make(map[string]struct{}),
	}
}

// Load 加载失败记录缓存
func (fc *FailedCache) Load(cacheFile string) error {
	file, err := os.Open(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("打开缓存文件失败: %w", err)
	}
	defer file.Close()

	fc.mu.Lock()
	defer fc.mu.Unlock()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ":") {
			fc.entries[line] = struct{}{}
		}
	}
	return nil
}

// Contains 检查账号密码是否已在失败缓存中（使用读锁，允许并发读）
func (fc *FailedCache) Contains(username, password string) bool {
	entry := username + ":" + password
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	_, ok := fc.entries[entry]
	return ok
}

// CountMisses 批量统计未命中缓存的任务数（单次加锁，避免 n×m 次锁竞争）
func (fc *FailedCache) CountMisses(users, passwords []string) int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	count := 0
	for _, u := range users {
		for _, p := range passwords {
			if _, ok := fc.entries[u+":"+p]; !ok {
				count++
			}
		}
	}
	return count
}

// Save 保存失败记录到缓存文件
func (fc *FailedCache) Save(cacheFile, username, password string) error {
	entry := username + ":" + password

	fc.mu.Lock()
	if _, ok := fc.entries[entry]; ok {
		fc.mu.Unlock()
		return nil
	}
	fc.entries[entry] = struct{}{}
	fc.mu.Unlock()

	writer, err := fc.getWriter(cacheFile)
	if err != nil {
		return err
	}

	if err := writer.Write(entry + "\n"); err != nil {
		logging.Warnf("写入失败缓存失败: %v", err)
	}
	return nil
}

// getWriter 获取或创建异步写入器
func (fc *FailedCache) getWriter(cacheFile string) (*filewriter.FileWriter, error) {
	fc.writerMu.Lock()
	defer fc.writerMu.Unlock()

	if fc.writer == nil {
		var err error
		fc.writer, err = filewriter.NewFileWriter(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("创建异步写入器失败: %w", err)
		}
	}

	return fc.writer, nil
}

// CloseWriter 关闭异步写入器（程序退出时调用）
func (fc *FailedCache) CloseWriter() error {
	fc.writerMu.Lock()
	defer fc.writerMu.Unlock()

	if fc.writer != nil {
		err := fc.writer.Close()
		fc.writer = nil
		return err
	}
	return nil
}
