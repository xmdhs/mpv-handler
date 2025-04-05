package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows/registry"
)

const protocol = "potplayer"

func registerProtocol() error {
	key, _, err := registry.CreateKey(registry.CLASSES_ROOT, protocol, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("创建注册表键失败: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue("", "URL:MPV Protocol"); err != nil {
		return fmt.Errorf("设置默认值失败: %w", err)
	}
	if err := key.SetStringValue("URL Protocol", ""); err != nil {
		return fmt.Errorf("设置 URL Protocol 失败: %w", err)
	}

	shellKey, _, err := registry.CreateKey(key, `shell\open\command`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("创建 shell\\open\\command 键失败: %w", err)
	}
	defer shellKey.Close()

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	command := fmt.Sprintf("\"%s\" \"%%1\"", exePath)
	if err := shellKey.SetStringValue("", command); err != nil {
		return fmt.Errorf("设置命令字符串失败: %w", err)
	}
	return nil
}

func processLink(link, mpvPath string) error {
	prefix := protocol + "://"
	if !strings.HasPrefix(link, prefix) {
		return fmt.Errorf("无效的协议链接: %s", link)
	}
	videoURL := strings.TrimPrefix(link, prefix)

	videoURL = strings.Replace(videoURL, "http//", "http://", 1)
	videoURL = strings.Replace(videoURL, "https//", "https://", 1)

	_, err := url.Parse(videoURL)
	if err != nil {
		return fmt.Errorf("url 错误: %w url: %v", err, videoURL)
	}
	cmd := exec.Command(mpvPath, videoURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动 mpv 失败: %w", err)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("未检测到链接参数，正在注册自定义协议到注册表中...")
		if err := registerProtocol(); err != nil {
			zenity.Error("注册协议失败，请尝试使用管理员运行\n\n"+err.Error(),
				zenity.Title("Error"),
				zenity.ErrorIcon)
			os.Exit(1)
		}
		zenity.Info("自定义协议注册成功！", zenity.InfoIcon, zenity.Title("成功"))
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		zenity.Error(err.Error(), zenity.Title("Error"), zenity.ErrorIcon)
		os.Exit(1)
	}
	dir, _ := filepath.Split(exePath)
	godotenv.Load(filepath.Join(dir, ".env"))

	mpvPath := os.Getenv("MPV_PATH")

	if mpvPath == "" {
		mpvPath = filepath.Join(dir, "mpv.exe")
	}

	link := os.Args[1]
	fmt.Printf("正在处理链接: %s\n", link)
	if err := processLink(link, mpvPath); err != nil {
		zenity.Error(err.Error(),
			zenity.Title("错误"),
			zenity.ErrorIcon)
		os.Exit(1)
	}
	fmt.Println("mpv 已成功启动。")
}
