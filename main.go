package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// 在文件开头添加颜色常量
const (
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
)

// 添加配置结构体
type Config struct {
	WaitTimeAfterReset     time.Duration
	CheckNetworkAfterReset bool
	AutoRestart            bool
	LogEnabled             bool
}

// 默认配置
var defaultConfig = Config{
	WaitTimeAfterReset:     5 * time.Second,
	CheckNetworkAfterReset: true,
	AutoRestart:            false,
	LogEnabled:             true,
}

func main() {
	// 设置控制台输出编码为UTF-8，以正确显示中文和emoji
	if runtime.GOOS == "windows" {
		// 启用 Windows 控制台的 ANSI 支持
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		proc := kernel32.NewProc("SetConsoleMode")
		handle, _, _ := proc.Call(uintptr(syscall.Stdout), 0x0001|0x0004)
		if handle == 0 {
			return
		}

		cmd := exec.Command("chcp", "65001")
		cmd.Run()
	}

	// 使用颜色输出标题
	fmt.Println(colorBold + "============================================" + colorReset)
	fmt.Println(colorCyan + colorBold + "          网络重置工具 v1.0" + colorReset)
	fmt.Println(colorBold + "============================================" + colorReset)
	fmt.Printf("%s系统类型:%s %s%s\n", colorBold, colorReset, colorYellow, runtime.GOOS+colorReset)
	fmt.Println(colorBold + "--------------------------------------------" + colorReset)
	fmt.Println(colorBlue + "🔍 正在检查管理员权限..." + colorReset)

	// 检查是否具有管理员权限
	if !checkAdminPrivileges() {
		fmt.Println("\n" + colorRed + "❌ 错误: 需要管理员权限!" + colorReset)
		fmt.Println("\n" + colorYellow + "📝 请按照以下步骤操作：" + colorReset)
		fmt.Println(colorBold + "--------------------------------------------" + colorReset)
		if runtime.GOOS == "windows" {
			fmt.Println("1️⃣  关闭当前窗口")
			fmt.Println("2️⃣  找到 NetReset.exe 程序")
			fmt.Println("3️⃣  右键点击该程序")
			fmt.Println("4️⃣  选择「以管理员身份运行」")
		} else {
			fmt.Println("1️⃣  关闭当前窗口")
			fmt.Println("2️⃣  打开终端")
			fmt.Println("3️⃣  输入: sudo ./NetReset")
		}
		fmt.Println(colorBold + "--------------------------------------------" + colorReset)
		fmt.Println("\n💡 按回车键退出...")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		return
	}

	fmt.Println(colorGreen + "✅ 权限检查通过!" + colorReset)
	fmt.Println(colorBold + "--------------------------------------------" + colorReset)
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
				fmt.Print(colorBlue + "正在创建桌面快捷方式..." + colorReset)
				if err := createShortcut(); err != nil {
					fmt.Printf("\n"+colorRed+"❌ %v\n"+colorReset, err)
				} else {
					fmt.Println(colorGreen + " ✅" + colorReset)
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
		fmt.Println(colorRed + "❌ 错误: 暂不支持当前操作系统" + colorReset)
		return
	}

	// 在程序结束前等待用户输入
	fmt.Println("\n💡 按回车键退出程序...")
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

// 添加一个统一的错误处理函数
func handleError(err error, message string) {
	if err != nil {
		fmt.Printf("\n"+colorRed+"❌ %s: %v"+colorReset+"\n", message, err)
	}
}

// 在执行命令时使用
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		handleError(err, fmt.Sprintf("执行命令失败: %s %s", name, strings.Join(args, " ")))
		fmt.Printf(colorRed+"📄 输出信息: %s\n"+colorReset, string(output))
		return err
	}
	return nil
}

func resetWindowsNetwork() {
	fmt.Println("\n" + colorCyan + "📝 开始执行网络重置..." + colorReset)
	fmt.Println(colorBold + "--------------------------------------------" + colorReset)

	steps := []struct {
		name    string
		command string
		args    []string
	}{
		{"重置WinHTTP代理设置", "netsh", []string{"winhttp", "reset", "proxy"}},
		{"清除IE代理服务器", "reg", []string{"delete", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyServer", "/f"}},
		{"禁用IE代理", "reg", []string{"delete", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyEnable", "/f"}},
		{"禁用LAN代理", "reg", []string{"add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f"}},
		{"启用自动检测设置", "reg", []string{"add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings\\Connections", "/v", "DefaultConnectionSettings", "/t", "REG_BINARY", "/d", "46000000090000000000000000000000000000000000000000", "/f"}},
		{"启用自动检测标志", "reg", []string{"add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "AutoDetect", "/t", "REG_DWORD", "/d", "1", "/f"}},
		{"删除自动配置脚本", "reg", []string{"delete", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "AutoConfigURL", "/f"}},
		{"刷新DNS缓存", "ipconfig", []string{"/flushdns"}},
		{"重置Winsock", "netsh", []string{"winsock", "reset"}},
		{"重置TCP/IP", "netsh", []string{"int", "ip", "reset"}},
	}

	successCount := 0
	totalSteps := len(steps)

	for _, step := range steps {
		fmt.Printf(colorBlue+"⏳ %s..."+colorReset, step.name)
		if err := execCommand(step.command, step.args...); err != nil {
			fmt.Println(colorRed + " ❌" + colorReset)
			continue
		}
		fmt.Println(colorGreen + " ✅" + colorReset)
		successCount++
	}

	fmt.Println(colorBold + "--------------------------------------------" + colorReset)
	fmt.Printf(colorCyan+"✨ 执行完成! 成功: %d/%d\n"+colorReset, successCount, totalSteps)
	fmt.Println("\n⚠️  重要提示：")
	fmt.Println("1. 建议重启电脑使所有设置生效")
	fmt.Println("2. 如果仍无法联网，请检查网线或WiFi连接")
	fmt.Println(colorBold + "--------------------------------------------" + colorReset)
	fmt.Println("\n💡 按回车键退出程序...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func resetMacNetwork() {
	fmt.Println("\n" + colorCyan + "📝 开始执行网络重置..." + colorReset)
	fmt.Println(colorBold + "--------------------------------------------" + colorReset)

	// 获取所有网络接口
	interfaces := []string{"Wi-Fi", "Ethernet", "USB 10/100/1000 LAN"}

	successCount := 0
	totalSteps := 0

	// 计算总步骤数
	for range interfaces {
		totalSteps += 3 // 每个接口有3个操作
	}
	totalSteps++ // 加上刷新DNS缓存的步骤

	for _, iface := range interfaces {
		steps := []struct {
			name    string
			command string
			args    []string
		}{
			{"关闭HTTP代理", "networksetup", []string{"-setwebproxystate", iface, "off"}},
			{"关闭HTTPS代理", "networksetup", []string{"-setsecurewebproxystate", iface, "off"}},
			{"关闭SOCKS代理", "networksetup", []string{"-setsocksfirewallproxystate", iface, "off"}},
		}

		for _, step := range steps {
			fmt.Printf(colorBlue+"⏳ [%s] %s..."+colorReset, iface, step.name)
			if err := execCommand(step.command, step.args...); err != nil {
				fmt.Println(colorRed + " ❌" + colorReset)
				continue
			}
			fmt.Println(colorGreen + " ✅" + colorReset)
			successCount++
		}
	}

	// 刷新DNS缓存
	fmt.Printf(colorBlue + "⏳ 刷新DNS缓存..." + colorReset)
	if err := execCommand("sudo", "killall", "-HUP", "mDNSResponder"); err != nil {
		fmt.Println(colorRed + " ❌" + colorReset)
	} else {
		fmt.Println(colorGreen + " ✅" + colorReset)
		successCount++
	}

	fmt.Println(colorBold + "--------------------------------------------" + colorReset)
	fmt.Printf(colorCyan+"✨ 执行完成! 成功: %d/%d\n"+colorReset, successCount, totalSteps)
	fmt.Println("\n⚠️  重要提示：")
	fmt.Println("1. 建议重启电脑使所有设置生效")
	fmt.Println("2. 如果仍无法联网，请检查网线或WiFi连接")
	fmt.Println(colorBold + "--------------------------------------------" + colorReset)
	fmt.Println("\n💡 按回车键退出程序...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
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

func checkNetworkConnection() bool {
	urls := []string{
		"https://www.baidu.com",
		"https://www.qq.com",
	}

	for _, url := range urls {
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		_, err := client.Get(url)
		if err == nil {
			return true
		}
	}
	return false
}

func writeLog(message string) error {
	logFile := "netReset.log"
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	return os.WriteFile(logFile, []byte(logMessage), os.ModeAppend)
}

func parseFlags() *Config {
	config := defaultConfig

	autoRestart := flag.Bool("restart", false, "重置后自动重启电脑")
	noWait := flag.Bool("nowait", false, "执行后立即退出")
	noLog := flag.Bool("nolog", false, "不记录日志")

	flag.Parse()

	config.AutoRestart = *autoRestart
	config.LogEnabled = !*noLog
	if *noWait {
		config.WaitTimeAfterReset = 0
	}

	return &config
}

func showProgress(current, total int) {
	width := 40
	percentage := float64(current) / float64(total)
	completed := int(percentage * float64(width))

	fmt.Printf("\r[")
	for i := 0; i < width; i++ {
		if i < completed {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.1f%%", percentage*100)
}
