package csvwriter

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewCSVWriter 创建写入器
func TestNewCSVWriter(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"Name", "Age", "City"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if w == nil {
		t.Fatal("写入器不应为 nil")
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}
}

// TestNewCSVWriterInvalidPath 无效路径
func TestNewCSVWriterInvalidPath(t *testing.T) {
	_, err := NewCSVWriter("/invalid/path/test.csv", []string{"A", "B"})
	if err == nil {
		t.Fatal("应返回错误")
	}
}

// TestWrite 写入单行
func TestCSVWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"Name", "Age", "City"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if err := w.Write([]string{"Alice", "30", "Beijing"}); err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Alice") || !strings.Contains(contentStr, "Beijing") {
		t.Fatalf("文件内容不完整: %s", contentStr)
	}
}

// TestWriteMultiple 写入多行
func TestCSVWriteMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"Name", "Age", "City"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	records := [][]string{
		{"Alice", "30", "Beijing"},
		{"Bob", "25", "Shanghai"},
		{"Charlie", "35", "Guangzhou"},
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatalf("写入失败: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	for _, record := range records {
		for _, field := range record {
			if !strings.Contains(contentStr, field) {
				t.Fatalf("文件内容不包含 '%s': %s", field, contentStr)
			}
		}
	}
}

// TestCSVHeaders 验证表头写入
func TestCSVHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"URL", "Username", "Password", "Status"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证表头
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	for _, header := range headers {
		if !strings.Contains(contentStr, header) {
			t.Fatalf("文件不包含表头 '%s': %s", header, contentStr)
		}
	}
}

// TestCSVAppend 追加写入
func TestCSVAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"Name", "Age"}

	// 第一次写入
	w1, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}
	if err := w1.Write([]string{"Alice", "30"}); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if err := w1.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 第二次写入（追加，不应重复写入表头）
	w2, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}
	if err := w2.Write([]string{"Bob", "25"}); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if err := w2.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Alice") || !strings.Contains(contentStr, "Bob") {
		t.Fatalf("文件内容不完整: %s", contentStr)
	}
}

// TestClose 关闭写入器
func TestCSVClose(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"Name"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}
}

// TestWriteConcurrent 并发写入
func TestCSVWriteConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"ID", "Value"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			record := []string{string(rune('0'+n)), "value" + string(rune('0'+n))}
			if err := w.Write(record); err != nil {
				t.Errorf("写入失败: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	for i := 0; i < 10; i++ {
		expected := "value" + string(rune('0'+i))
		if !strings.Contains(contentStr, expected) {
			t.Fatalf("文件内容不包含 '%s': %s", expected, contentStr)
		}
	}
}

// TestCSVReadBack 读取验证 CSV 格式
func TestCSVReadBack(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")
	headers := []string{"Name", "Age", "City"}

	w, err := NewCSVWriter(filePath, headers)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	records := [][]string{
		{"Alice", "30", "Beijing"},
		{"Bob", "25", "Shanghai"},
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatalf("写入失败: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 使用 csv.Reader 读取验证格式
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("打开文件失败: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	allRecords, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("读取 CSV 失败: %v", err)
	}

	// 验证记录数（表头 + 2 条数据）
	if len(allRecords) != 3 {
		t.Fatalf("期望 3 条记录，实际 %d 条", len(allRecords))
	}

	// 验证表头
	for i, h := range headers {
		if allRecords[0][i] != h {
			t.Fatalf("表头不匹配: 期望 '%s', 实际 '%s'", h, allRecords[0][i])
		}
	}

	// 验证数据
	for i, record := range records {
		for j, field := range record {
			if allRecords[i+1][j] != field {
				t.Fatalf("记录 %d 字段 %d 不匹配: 期望 '%s', 实际 '%s'", i, j, field, allRecords[i+1][j])
			}
		}
	}
}
