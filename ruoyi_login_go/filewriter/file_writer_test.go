package filewriter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewFileWriter 创建写入器
func TestNewFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	fw, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if fw == nil {
		t.Fatal("写入器不应为 nil")
	}

	if err := fw.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}
}

// TestNewFileWriterInvalidPath 无效路径
func TestNewFileWriterInvalidPath(t *testing.T) {
	_, err := NewFileWriter("/invalid/path/test.txt")
	if err == nil {
		t.Fatal("应返回错误")
	}
}

// TestWrite 写入单行
func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	fw, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if err := fw.Write("hello world\n"); err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	if err := fw.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	if !strings.Contains(string(content), "hello world") {
		t.Fatalf("文件内容不包含 'hello world': %s", string(content))
	}
}

// TestWriteMultiple 写入多行
func TestWriteMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	fw, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	lines := []string{"line1\n", "line2\n", "line3\n"}
	for _, line := range lines {
		if err := fw.Write(line); err != nil {
			t.Fatalf("写入失败: %v", err)
		}
	}

	if err := fw.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	for _, line := range lines {
		if !strings.Contains(string(content), strings.TrimSuffix(line, "\n")) {
			t.Fatalf("文件内容不包含 '%s': %s", line, string(content))
		}
	}
}

// TestWriteAppend 追加写入
func TestWriteAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// 第一次写入
	fw1, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}
	if err := fw1.Write("first\n"); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if err := fw1.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 第二次写入（追加）
	fw2, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}
	if err := fw2.Write("second\n"); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if err := fw2.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "first") || !strings.Contains(contentStr, "second") {
		t.Fatalf("文件内容不完整: %s", contentStr)
	}
}

// TestClose 关闭写入器
func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	fw, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	if err := fw.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}
}

// TestWriteConcurrent 并发写入
func TestWriteConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	fw, err := NewFileWriter(filePath)
	if err != nil {
		t.Fatalf("创建写入器失败: %v", err)
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			line := "concurrent line " + string(rune('0'+n)) + "\n"
			if err := fw.Write(line); err != nil {
				t.Errorf("写入失败: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	if err := fw.Close(); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	contentStr := string(content)
	for i := 0; i < 10; i++ {
		expected := "concurrent line " + string(rune('0'+i))
		if !strings.Contains(contentStr, expected) {
			t.Fatalf("文件内容不包含 '%s': %s", expected, contentStr)
		}
	}
}
