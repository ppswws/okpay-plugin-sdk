package host

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pluginIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func ensureInsideDir(path, base string) error {
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return fmt.Errorf("无法解析目录: %w", err)
	}
	targetAbs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("无法解析路径: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(baseAbs); err == nil {
		baseAbs = resolved
	}
	if resolved, err := filepath.EvalSymlinks(targetAbs); err == nil {
		targetAbs = resolved
	}
	if targetAbs == baseAbs {
		return fmt.Errorf("路径非法: %s", targetAbs)
	}
	if !strings.HasPrefix(targetAbs, baseAbs+string(os.PathSeparator)) {
		return fmt.Errorf("路径不在插件目录内: %s", targetAbs)
	}
	return nil
}

func ensurePluginPath(path, base string) error {
	if err := ensureInsideDir(path, base); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("无法访问插件文件: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("禁止加载符号链接插件: %w", fmt.Errorf("forbid symlink"))
	}
	if info.IsDir() {
		return fmt.Errorf("路径是目录: %s", path)
	}
	if info.Mode()&0o111 == 0 {
		// 尝试补齐执行权限，避免上传时因权限位缺失导致失败。
		if err := os.Chmod(path, info.Mode()|0o750); err != nil {
			return fmt.Errorf("插件文件未设置可执行权限: %w", err)
		}
		if refreshed, err := os.Lstat(path); err == nil {
			if refreshed.Mode()&0o111 == 0 {
				return fmt.Errorf("插件文件未设置可执行权限: %w", fmt.Errorf("mode %v", refreshed.Mode()))
			}
		}
	}
	return nil
}

func ensureRegularFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("无法访问插件文件: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("禁止加载符号链接插件: %w", fmt.Errorf("forbid symlink"))
	}
	if info.IsDir() {
		return fmt.Errorf("路径是目录: %s", path)
	}
	return nil
}

func validatePluginID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("插件 ID 不能为空")
	}
	if !pluginIDPattern.MatchString(id) {
		return fmt.Errorf("插件 ID 只允许字母、数字、下划线或中划线")
	}
	return nil
}
