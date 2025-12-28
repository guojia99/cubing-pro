package wca

import (
	"archive/zip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// copyFile 复制文件（用于跨文件系统的情况）
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// checkRemoteFileDate 从重定向后的 URL 中提取日期（YYYYMMDD）
func checkRemoteFileDate() (ts time.Time, url string, err error) {
	client := &http.Client{}
	resp, err := client.Get(syncUrl)
	if err != nil {
		return time.Time{}, "", err
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	parts := strings.Split(finalURL, "/")
	filename := parts[len(parts)-1]
	re := regexp.MustCompile(`(\d{8}T\d{6}Z)`)

	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return time.Time{}, finalURL, fmt.Errorf("no timestamp found in URL")
	}
	timestampStr := matches[1]
	ts, err = time.Parse("20060102T150405Z", timestampStr)

	return ts, finalURL, err
}

// downloadIfNeeded 下载文件（如果当天的 zip 不存在）
// - 使用当前系统时间生成目标文件名：YYYYMMDD.zip
// - url 仅用于下载，不做任何解析或校验
func downloadIfNeeded(targetDir, fileURL string) (targetFile string, err error) {
	dateStr := time.Now().Format("20060102") // e.g., "20251223"
	targetFile = filepath.Join(targetDir, dateStr+".zip")
	if _, err = os.Stat(targetFile); err == nil {
		return
	}

	_ = os.MkdirAll(targetDir, 0755)

	// 4. 临时文件下载
	tmpFile, err := os.CreateTemp("/tmp", "download_*.zip")
	if err != nil {
		return
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 下载文件
	resp, err := http.Get(fileURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP error: %s", resp.Status)
		return
	}

	// 写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return
	}

	if copyErr := copyFile(tmpFile.Name(), targetFile); copyErr != nil {
		return
	}
	return
}

// extractFile 解压单个 ZIP 条目到目标目录
func extractFile(f *zip.File, targetDir string) error {
	// 检查是否为目录（有些 ZIP 条目是目录）
	if f.FileInfo().IsDir() {
		dirPath := filepath.Join(targetDir, f.Name)
		return os.MkdirAll(dirPath, f.Mode())
	}

	// 安全性：防止路径穿越（例如 ../../etc/passwd）
	if containsDotDot(f.Name) {
		return fmt.Errorf("invalid file path in zip (contains '..'): %s", f.Name)
	}

	// 构造完整输出路径
	outPath := filepath.Join(targetDir, f.Name)

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	// 打开 ZIP 中的文件
	inFile, err := f.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()

	// 创建目标文件
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 复制内容
	_, err = io.Copy(outFile, inFile)
	return err
}

// containsDotDot 检查路径是否包含 ".."（防止 zip slip 攻击）
func containsDotDot(v string) bool {
	if !filepath.IsAbs(v) {
		v = filepath.Clean(v)
	}
	return filepath.HasPrefix(v, "..") || v == ".."
}

// extractZipToDb 解压 zip 文件到 dbPath/YYYYMMDD/
func extractZipToDb(fileName, dbPath string) (targetDir string, err error) {
	// 1. 从 fileName 提取日期（如 /data/wca/20251223.zip → 20251223）
	base := filepath.Base(fileName)
	re := regexp.MustCompile(`^(\d{8})\.zip$`)
	matches := re.FindStringSubmatch(base)
	if len(matches) < 2 {
		err = fmt.Errorf("filename does not match YYYYMMDD.zip pattern: %s", fileName)
		return
	}
	dateStr := matches[1]

	// 2. 构造目标解压目录
	targetDir = filepath.Join(dbPath, dateStr)

	// 3. 如果目录已存在，可选择跳过或覆盖。这里我们先清理（可选）
	if _, err = os.Stat(targetDir); err == nil {
		_ = os.RemoveAll(targetDir)
	}

	// 4. 创建目标目录
	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return
	}

	// 5. 打开 ZIP 文件
	r, err := zip.OpenReader(fileName)
	if err != nil {
		return
	}
	defer r.Close()

	// 6. 遍历 ZIP 中的每个文件并解压
	for _, f := range r.File {
		if err = extractFile(f, targetDir); err != nil {
			err = fmt.Errorf("failed to extract file %s: %w", f.Name, err)
			return
		}
	}

	fmt.Printf("Successfully extracted %s to %s\n", fileName, targetDir)
	return
}

func isSameUTCDay(t1, t2 time.Time) bool {
	return t1.UTC().Truncate(24 * time.Hour).Equal(t2.UTC().Truncate(24 * time.Hour))
}

// parseMySQLDSN 解析 DSN 并返回 user, pass, host, port
func parseMySQLDSN(dsn string) (user, pass, host, port string, err error) {
	// 移除可选的数据库名和查询参数
	dsn = strings.Split(dsn, "?")[0]   // 去掉 ?charset=...
	dsn = strings.TrimSuffix(dsn, "/") // 去掉末尾 /

	if !strings.Contains(dsn, "@") {
		return "", "", "", "", fmt.Errorf("invalid DSN: missing '@'")
	}

	parts := strings.SplitN(dsn, "@", 2)
	authPart := parts[0]
	addrPart := parts[1]

	// 解析 user[:pass]
	user = authPart
	pass = ""
	if idx := strings.Index(authPart, ":"); idx != -1 {
		user = authPart[:idx]
		pass = authPart[idx+1:]
	}

	// 解析地址部分
	host = "localhost"
	port = "3306" // 默认

	if strings.HasPrefix(addrPart, "tcp(") && strings.HasSuffix(addrPart, ")") {
		addr := addrPart[4 : len(addrPart)-1] // 去掉 tcp(...)
		if h, p, err := net.SplitHostPort(addr); err == nil {
			host = h
			port = p
		} else {
			// 可能只有 host（无端口）
			host = addr
		}
	} else if strings.HasPrefix(addrPart, "unix(") {
		return "", "", "", "", fmt.Errorf("unix socket not supported in importSQLFile")
	} else {
		// 裸地址（不推荐）
		if h, p, err := net.SplitHostPort(addrPart); err == nil {
			host = h
			port = p
		} else {
			host = addrPart
		}
	}

	return user, pass, host, port, nil
}

// importSQLFileViaShell 使用 shell 执行 mysql ... < file.sql 命令
func importSQLFileViaShell(dbName, sqlFile, dsn string) error {
	user, pass, host, port, err := parseMySQLDSN(dsn)
	if err != nil {
		return fmt.Errorf("解析 DSN 失败: %w", err)
	}

	// 检查 SQL 文件是否存在
	if _, err = os.Stat(sqlFile); os.IsNotExist(err) {
		return fmt.Errorf("SQL 文件不存在: %s", sqlFile)
	}

	// 构造完整的 shell 命令字符串
	// 注意：密码中若含特殊字符（如 $、!、空格等）需转义，此处为简化未处理
	var ps = ""
	if pass != "" {
		ps = fmt.Sprintf("-p%s", pass)
	}

	cmdStr := fmt.Sprintf(
		"mysql -u %s %s -h %s -P %s %s < %s",
		user, ps, host, port, dbName, sqlFile,
	)

	fmt.Printf("执行命令: %s\n", cmdStr) // 可选：用于调试（注意不要泄露密码）

	// 通过 /bin/sh 执行命令（Linux/macOS）
	cmd := exec.Command("/bin/sh", "-c", cmdStr)

	// 可选：捕获输出以便查看错误
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shell 命令执行失败: %w\n输出: %s", err, string(output))
	}

	// 成功时 output 通常为空（除非 mysql 打印了警告）
	if len(output) > 0 {
		fmt.Printf("命令输出: %s\n", string(output))
	}

	return nil
}

// 辅助函数
func isDigitsOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// cleanSQLWithSed 执行: sed -i 's|/\*M![^*]*\*/||g' <sqlFile>
func cleanSQLWithSed(sqlFile string) error {
	cmd := exec.Command("sed", "-i", `s|/\*M![^*]*\*/||g`, sqlFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sed command failed: %w, output: %s", err, string(out))
	}
	return nil
}
