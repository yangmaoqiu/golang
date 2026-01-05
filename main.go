package main

import (
	"flag"
	"fmt"
	"runtime"
)

func main() {
	// 定义命令行参数
	testMode := flag.Bool("test", false, "运行测试模式")
	platformTest := flag.Bool("platform", false, "运行平台测试")
	unixTest := flag.Bool("unix-test", false, "运行 Unix 平台测试（仅 Unix 系统）")
	readTest := flag.Bool("read-test", false, "运行读取文件参数测试")
	help := flag.Bool("h", false, "显示帮助信息")
	helpLong := flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	// 显示帮助信息
	if *help || *helpLong {
		printHelp()
		return
	}

	// 运行平台测试
	if *platformTest {
		fmt.Println("=== 运行平台测试 ===")
		TestMain()
		return
	}

	// 运行 Unix 测试
	if *unixTest {
		fmt.Println("=== 运行 Unix 平台测试 ===")
		fmt.Println("注意: 此功能仅在 Unix 系统（Linux/macOS）上可用")
		fmt.Println("当前系统:", runtime.GOOS)
		return
	}

	// 运行读取文件测试
	if *readTest {
		fmt.Println("=== 运行读取文件参数测试 ===")
		testReadFileVariations()
		return
	}

	// 运行测试模式
	if *testMode {
		fmt.Println("=== 运行测试模式 ===")
		test()
		return
	}

	// 默认运行主测试
	fmt.Println("=== Rockey-ARM 测试程序 ===")
	fmt.Println("使用以下命令行参数:")
	fmt.Println("  -test      运行设备测试")
	fmt.Println("  -platform  运行平台测试")
	fmt.Println("  -unix-test 运行 Unix 平台测试（仅 Unix 系统）")
	fmt.Println("  -read-test 运行读取文件参数测试")
	fmt.Println("  -h, -help 显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run . -test")
	fmt.Println("  go run . -platform")
}

func printHelp() {
	fmt.Println("Rockey-ARM 测试程序")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  test-go [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -test      运行设备测试（测试加密狗功能）")
	fmt.Println("  -platform  运行平台测试（测试跨平台兼容性）")
	fmt.Println("  -unix-test 运行 Unix 平台测试（仅 Unix 系统）")
	fmt.Println("  -read-test 运行读取文件参数测试（测试不同参数组合）")
	fmt.Println("  -h, -help 显示帮助信息")
	fmt.Println()
	fmt.Println("描述:")
	fmt.Println("  这是一个跨平台的 Rockey-ARM 加密狗测试程序。")
	fmt.Println("  支持 Windows、Linux 和 macOS 系统。")
	fmt.Println()
	fmt.Println("库文件位置:")
	fmt.Println("  Windows:    lib/windows/Dongle_d.dll (64位)")
	fmt.Println("              lib/windows/Dongle_d32.dll (32位)")
	fmt.Println("  Linux:      lib/linux/arm64/libRockeyARM.so.0.3 (ARM64)")
	fmt.Println("              lib/linux/loong64/libRockeyARM.so.0.3 (LoongArch64)")
	fmt.Println("  macOS:      lib/darwin/libRockeyARM.dylib")
	fmt.Println()
	fmt.Println("构建:")
	fmt.Println("  go build -o rockey-test.exe  # Windows")
	fmt.Println("  go build -o rockey-test      # Linux/macOS")
}
