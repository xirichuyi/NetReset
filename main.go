package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	// 设置控制台输出编码为UTF-8，以正确显示中文和emoji
	if runtime.GOOS == "windows" {
		cmd := exec.Command("chcp", "65001")
		cmd.Run()
	}

	fmt.Printf("=== 网络重置工具 v1.0 ===\n")
	fmt.Printf("当前操作系统: %s\n", runtime.GOOS)
	fmt.Println("开始检查系统权限...")

	// 检查是否具有管理员权限
	if !checkAdminPrivileges() {
		fmt.Println("\n❌ 错误: 没有管理员权限!")
		fmt.Println("\n请按照以下步骤操作：")
		if runtime.GOOS == "windows" {
			fmt.Println("\n1. 关闭当前窗口")
			fmt.Println("2. 在文件夹中找到 NetReset.exe")
			fmt.Println("3. 右键点击 NetReset.exe")
			fmt.Println("4. 在弹出菜单中选择「以管理员身份运行」")
		} else {
			fmt.Println("\n1. 关闭当前窗口")
			fmt.Println("2. 打开终端")
			fmt.Println("3. 输入: sudo ./NetReset")
		}

		fmt.Println("\n按回车键退出程序...")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		return
	}

	fmt.Println("✅ 权限检查通过")
	fmt.Println("\n开始重置网络设置...")

	// 在权限检查通过后添加创建快捷方式的选项
	if runtime.GOOS == "windows" {
		// 检查快捷方式是否已存在
		if !shortcutExists() {
			fmt.Println("\n是否创建桌面快捷方式？(Y/N)")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "y" || answer == "yes" {
				fmt.Print("正在创建桌面快捷方式...")
				if err := createShortcut(); err != nil {
					fmt.Printf("\n❌ %v\n", err)
				} else {
					fmt.Println(" ✅")
					fmt.Println("\n提示：桌面快捷方式已创建，双击使用时请选择「以管理员身份运行」")
				}
			}
		}
	}

	switch runtime.GOOS {
	case "windows":
		resetWindowsNetwork()
	case "darwin":
		resetMacNetwork()
	default:
		fmt.Println("❌ 错误: 暂不支持当前操作系统")
		return
	}

	// 在程序结束前等待用户输入
	fmt.Println("\n按回车键退出程序...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func checkAdminPrivileges() bool {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	} else {
		cmd := exec.Command("id", "-u")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(output)) == "0"
	}
}

func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("❌ 执行命令失败: %s %s\n", name, strings.Join(args, " "))
		fmt.Printf("错误信息: %s\n", string(output))
		return err
	}
	return nil
}

func resetWindowsNetwork() {
	steps := []struct {
		name    string
		command string
		args    []string
	}{
		{"重置WinHTTP代理设置", "netsh", []string{"winhttp", "reset", "proxy"}},
		{"清除IE代理服务器", "reg", []string{"delete", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyServer", "/f"}},
		{"禁用IE代理", "reg", []string{"delete", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyEnable", "/f"}},
		{"刷新DNS缓存", "ipconfig", []string{"/flushdns"}},
		{"重置Winsock", "netsh", []string{"winsock", "reset"}},
		{"重置TCP/IP", "netsh", []string{"int", "ip", "reset"}},
	}

	for _, step := range steps {
		fmt.Printf("\n🔄 %s...", step.name)
		if err := execCommand(step.command, step.args...); err != nil {
			continue
		}
		fmt.Printf(" ✅")
	}

	fmt.Println("\n\n✅ Windows网络设置重置完成!")
	fmt.Println("\n⚠️  建议重启电脑使所有设置生效")
	fmt.Println("⏳ 程序将在5秒后自动关闭...")
	time.Sleep(5 * time.Second)
}

func resetMacNetwork() {
	steps := []struct {
		name    string
		command string
		args    []string
	}{
		{"关闭HTTP代理", "networksetup", []string{"-setwebproxystate", "Wi-Fi", "off"}},
		{"关闭HTTPS代理", "networksetup", []string{"-setsecurewebproxystate", "Wi-Fi", "off"}},
		{"关闭SOCKS代理", "networksetup", []string{"-setsocksfirewallproxystate", "Wi-Fi", "off"}},
		{"刷新DNS缓存", "killall", []string{"-HUP", "mDNSResponder"}},
	}

	for _, step := range steps {
		fmt.Printf("\n🔄 %s...", step.name)
		if err := execCommand(step.command, step.args...); err != nil {
			continue
		}
		fmt.Printf(" ✅")
	}

	fmt.Println("\n\n✅ macOS网络设置重置完成!")
	fmt.Println("\n⚠️  如果仍然无法访问网络，请尝试重启电脑")
	fmt.Println("⏳ 程序将在5秒后自动关闭...")
	time.Sleep(5 * time.Second)
}

func createShortcut() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("仅支持在 Windows 系统创建快捷方式")
	}

	// 获取当前可执行文件的完整路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}

	// 获取桌面路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %v", err)
	}
	desktopPath := filepath.Join(homeDir, "Desktop")

	// 创建快捷方式的 VBS 脚本
	vbsContent := fmt.Sprintf(`
Set ws = CreateObject("WScript.Shell")
Set shortcut = ws.CreateShortcut("%s\NetReset.lnk")
shortcut.TargetPath = "%s"
shortcut.WorkingDirectory = "%s"
shortcut.Description = "网络重置工具"
shortcut.IconLocation = "%s"
shortcut.Arguments = ""
shortcut.WindowStyle = 1
shortcut.Save
`, desktopPath, exePath, filepath.Dir(exePath), exePath)

	// 创建临时 VBS 文件
	tmpFile := filepath.Join(os.TempDir(), "create_shortcut.vbs")
	if err := os.WriteFile(tmpFile, []byte(vbsContent), 0644); err != nil {
		return fmt.Errorf("创建脚本文件失败: %v", err)
	}
	defer os.Remove(tmpFile)

	// 执行 VBS 脚本
	cmd := exec.Command("cscript", "//Nologo", tmpFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建快捷方式失败: %v", err)
	}

	return nil
}

// 添加新函数：检查快捷方式是否存在
func shortcutExists() bool {
	// 获取桌面路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// 检查中文和英文桌面文件夹
	desktopPaths := []string{
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "桌面"),
	}

	shortcutName := "NetReset.lnk"

	// 检查快捷方式是否存在
	for _, desktopPath := range desktopPaths {
		shortcutPath := filepath.Join(desktopPath, shortcutName)
		if _, err := os.Stat(shortcutPath); err == nil {
			return true
		}
	}

	return false
}
