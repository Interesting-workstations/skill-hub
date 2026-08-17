// Package config 提供轻量的 .env 文件加载（无第三方依赖）。
package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv 从指定路径加载 .env 文件到进程环境变量。
// 已存在的环境变量不会被覆盖；文件不存在时静默跳过。
func LoadEnv(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行与注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// EnvOr 读取环境变量，未设置时返回默认值。
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// EnvFile 返回当前运行环境对应的 env 文件路径。
// 通过 SKILLHUB_ENV 环境变量区分：prod/production/1 → .env.prod（线上），
// 其他（默认，本地开发）→ .env。
func EnvFile() string {
	switch strings.ToLower(os.Getenv("SKILLHUB_ENV")) {
	case "prod", "production", "1":
		return ".env.prod"
	default:
		return ".env"
	}
}
