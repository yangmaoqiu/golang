package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// ============ 常量定义 ============

// 错误码定义
const (
	DONGLE_SUCCESS           = 0x00000000 // 成功
	DONGLE_NOT_FOUND         = 0xF0000001 // 未找到设备
	DONGLE_INVALID_HANDLE    = 0xF0000002 // 无效句柄
	DONGLE_INVALID_PARAMETER = 0xF0000003 // 无效参数
	DONGLE_ACCESS_DENIED     = 0xF0000004 // 访问被拒绝
	DONGLE_NO_MORE_DEVICE    = 0xF0000005 // 没有更多设备
	DONGLE_NEED_FIND         = 0xF0000006 // 需要查找设备
	DONGLE_INVALID_PASSWORD  = 0xF0000007 // 无效密码
	DONGLE_INVALID_DEVID     = 0xF0000008 // 无效设备ID
	DONGLE_INVALID_BUFFER    = 0xF0000009 // 无效缓冲区
	DONGLE_INVALID_FILEID    = 0xF000000A // 无效文件ID
	DONGLE_INVALID_OFFSET    = 0xF000000B // 无效偏移量
	DONGLE_INVALID_SIZE      = 0xF000000C // 无效大小
	DONGLE_UNKNOWN_ERROR     = 0xFFFFFFFF // 未知错误
)

// 函数名称常量
const (
	FUNC_ENUM     = "Dongle_Enum"
	FUNC_OPEN     = "Dongle_Open"
	FUNC_READFILE = "Dongle_ReadFile"
	FUNC_CLOSE    = "Dongle_Close"
)

// 测试常量
const (
	TEST_BUFFER_SIZE = 1024   // 测试缓冲区大小
	TEST_FILE_ID     = 0x0001 // 测试文件ID
	TEST_OFFSET      = 0      // 测试偏移量
)

// ============ 结构体定义 ============

// DongleInfo 设备信息结构体
type DongleInfo struct {
	MVer      uint16  // 版本号
	MType     uint16  // 类型
	MBirthDay [8]byte // 生产日期
	MAgent    uint32  // 代理商ID
	MPID      uint32  // 产品ID
	MUserID   uint32  // 用户ID
	MHID      [8]byte // 硬件ID
	MIsMother uint32  // 是否母锁
	MDevType  uint32  // 设备类型
}

// DongleHandle 设备句柄
type DongleHandle uintptr

// ============ 全局变量 ============

var (
	// 命令行参数
	testMode     = flag.Bool("test", false, "运行设备测试")
	platformTest = flag.Bool("platform", false, "运行平台测试")
	readTest     = flag.Bool("read-test", false, "运行读取文件参数测试")
	help         = flag.Bool("h", false, "显示帮助信息")
	helpLong     = flag.Bool("help", false, "显示帮助信息")
	diagnoseMode = flag.Bool("diagnose", false, "运行详细诊断模式")
)

// ============ 辅助函数 ============

// getLibraryPath 根据系统架构返回库文件路径
func getLibraryPath() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	switch runtime.GOARCH {
	case "arm64", "aarch64":
		return "./lib/linux/arm64/libRockeyARM.so.0.3"
	case "loong64":
		return "./lib/linux/loong64/libRockeyARM.so.0.3"
	case "amd64", "x86_64":
		return "./lib/linux/libRockeyARM.so"
	default:
		return "./lib/linux/libRockeyARM.so"
	}
}

// loadLibrary 加载动态库
func loadLibrary(libPath string) (uintptr, error) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(libPath)
	if err != nil {
		return 0, fmt.Errorf("库文件不存在: %v", err)
	}

	fmt.Printf("  库文件信息: 大小=%d字节, 权限=%v\n", fileInfo.Size(), fileInfo.Mode())

	// 使用 purego 加载库
	handle, err := purego.Dlopen(libPath, purego.RTLD_LAZY)
	if err != nil {
		return 0, fmt.Errorf("加载库失败: %v", err)
	}

	fmt.Printf("  库句柄: 0x%x\n", handle)

	return handle, nil
}

// getProcAddress 获取函数地址
func getProcAddress(handle uintptr, funcName string) (uintptr, error) {
	// 使用 purego 获取函数地址
	addr, err := purego.Dlsym(handle, funcName)
	if err != nil {
		// 尝试带下划线的版本
		underscoreName := "_" + funcName
		addr, err = purego.Dlsym(handle, underscoreName)
		if err != nil {
			return 0, fmt.Errorf("找不到函数 %s: %v (尝试了 %s 和 %s)", funcName, err, funcName, underscoreName)
		}
		fmt.Printf("  函数 %s 找到 (使用带下划线版本: %s)\n", funcName, underscoreName)
	} else {
		fmt.Printf("  函数 %s 找到\n", funcName)
	}
	fmt.Printf("  函数地址: 0x%x\n", addr)
	return addr, nil
}

// getErrorDescription 获取错误码描述
func getErrorDescription(errorCode uint32) string {
	switch errorCode {
	case DONGLE_SUCCESS:
		return "成功"
	case DONGLE_NOT_FOUND:
		return "未找到设备"
	case DONGLE_INVALID_HANDLE:
		return "无效句柄"
	case DONGLE_INVALID_PARAMETER:
		return "无效参数"
	case DONGLE_ACCESS_DENIED:
		return "访问被拒绝"
	case DONGLE_NO_MORE_DEVICE:
		return "没有更多设备"
	case DONGLE_NEED_FIND:
		return "需要查找设备"
	case DONGLE_INVALID_PASSWORD:
		return "无效密码"
	case DONGLE_INVALID_DEVID:
		return "无效设备ID"
	case DONGLE_INVALID_BUFFER:
		return "无效缓冲区"
	case DONGLE_INVALID_FILEID:
		return "无效文件ID"
	case DONGLE_INVALID_OFFSET:
		return "无效偏移量"
	case DONGLE_INVALID_SIZE:
		return "无效大小"
	case DONGLE_UNKNOWN_ERROR:
		return "未知错误"
	default:
		return fmt.Sprintf("未知错误码: %08X", errorCode)
	}
}

// showBinHex 显示二进制数据的十六进制格式
func showBinHex(data []byte) {
	// 每行显示 16 字节
	for i := 0; i < len(data); i += 16 {
		// 显示十六进制
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Printf("%02X ", data[i+j])
			} else {
				fmt.Printf("   ") // 对齐用空格
			}

			if j == 7 {
				fmt.Printf("- ")
			}
		}

		fmt.Printf("    ")

		// 显示 ASCII 字符
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Printf(".")
			}
		}

		fmt.Println()
	}
	fmt.Println()
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============ 核心功能函数 ============

// enumDevices 枚举设备
func enumDevices(enumFunc uintptr) ([]DongleInfo, int, uint32, error) {
	var countLocal int32

	fmt.Printf("  调用枚举函数，地址: 0x%x\n", enumFunc)

	// 使用 purego 的正确方式调用函数
	// 定义函数原型
	type EnumFuncType func(infoList unsafe.Pointer, count *int32) uint32

	// 将函数指针转换为可调用的函数
	var enumFuncGo EnumFuncType
	purego.RegisterFunc(&enumFuncGo, enumFunc)

	// 第一次调用获取设备数量
	fmt.Printf("  第一次调用: 获取设备数量\n")
	fmt.Printf("  参数: infoList=0 (NULL), count=%p\n", unsafe.Pointer(&countLocal))

	retCode := enumFuncGo(nil, &countLocal)
	fmt.Printf("  返回码: 0x%08X (%s)\n", retCode, getErrorDescription(retCode))

	if retCode != DONGLE_SUCCESS {
		return nil, 0, retCode, fmt.Errorf(getErrorDescription(retCode))
	}

	fmt.Printf("  找到设备数量: %d\n", countLocal)

	if countLocal == 0 {
		return nil, 0, DONGLE_NOT_FOUND, fmt.Errorf(getErrorDescription(DONGLE_NOT_FOUND))
	}

	// 分配内存存储设备信息
	count := int(countLocal)
	keyList := make([]DongleInfo, count)

	fmt.Printf("  分配设备信息数组: %d 个元素\n", count)

	// 第二次调用获取详细信息
	fmt.Printf("  第二次调用: 获取设备详细信息\n")
	fmt.Printf("  参数: infoList=%p, count=%p\n", unsafe.Pointer(&keyList[0]), unsafe.Pointer(&countLocal))

	retCode = enumFuncGo(unsafe.Pointer(&keyList[0]), &countLocal)
	fmt.Printf("  返回码: 0x%08X (%s)\n", retCode, getErrorDescription(retCode))

	if retCode != DONGLE_SUCCESS {
		return nil, 0, retCode, fmt.Errorf(getErrorDescription(retCode))
	}

	fmt.Printf("  设备信息获取成功\n")
	return keyList, count, retCode, nil
}

// openDevice 打开设备
func openDevice(openFunc uintptr, index int) (DongleHandle, uint32, error) {
	var hKeyLocal DongleHandle

	fmt.Printf("  调用打开函数，地址: 0x%x\n", openFunc)
	fmt.Printf("  参数: handle=%p, index=%d\n", unsafe.Pointer(&hKeyLocal), index)

	// 使用 purego 的正确方式调用函数
	// 定义函数原型
	type OpenFuncType func(handle *DongleHandle, index int) uint32

	// 将函数指针转换为可调用的函数
	var openFuncGo OpenFuncType
	purego.RegisterFunc(&openFuncGo, openFunc)

	retCode := openFuncGo(&hKeyLocal, index)
	fmt.Printf("  返回码: 0x%08X (%s)\n", retCode, getErrorDescription(retCode))

	if retCode != DONGLE_SUCCESS {
		return 0, retCode, fmt.Errorf(getErrorDescription(retCode))
	}

	fmt.Printf("  设备句柄: 0x%x\n", hKeyLocal)
	return hKeyLocal, retCode, nil
}

// readFile 读取文件
func readFile(readFileFunc uintptr, handle DongleHandle, fileID, offset uintptr, buffer []byte) (uint32, int, error) {
	if handle == 0 {
		return DONGLE_INVALID_HANDLE, 0, fmt.Errorf(getErrorDescription(DONGLE_INVALID_HANDLE))
	}

	if len(buffer) == 0 {
		return DONGLE_INVALID_BUFFER, 0, fmt.Errorf(getErrorDescription(DONGLE_INVALID_BUFFER))
	}

	fmt.Printf("  调用读取文件函数，地址: 0x%x\n", readFileFunc)
	fmt.Printf("  参数: handle=0x%x, fileID=0x%x, offset=0x%x, buffer=%p, size=%d\n",
		handle, fileID, offset, unsafe.Pointer(&buffer[0]), len(buffer))

	// 使用 purego 的正确方式调用函数
	// 尝试不同的函数原型
	// 方式1: 5个参数的函数原型（不包含文件类型）
	type ReadFileFuncType1 func(handle DongleHandle, fileID uintptr, offset uintptr, buffer unsafe.Pointer, size uintptr) uint32

	// 方式2: 6个参数的函数原型（包含文件类型）
	type ReadFileFuncType2 func(handle DongleHandle, fileType uintptr, fileID uintptr, offset uintptr, buffer unsafe.Pointer, size uintptr) uint32

	// 先尝试方式1（5个参数）
	var readFileFuncGo1 ReadFileFuncType1
	purego.RegisterFunc(&readFileFuncGo1, readFileFunc)

	retCode := readFileFuncGo1(handle, fileID, offset, unsafe.Pointer(&buffer[0]), uintptr(len(buffer)))
	fmt.Printf("  调用结果 (5参数): 返回码=0x%08X (%s)\n", retCode, getErrorDescription(retCode))

	if retCode == DONGLE_SUCCESS {
		return retCode, len(buffer), nil
	}

	// 如果方式1失败，尝试方式2（6个参数）
	fmt.Printf("  尝试方式2: 6个参数（包含文件类型）\n")
	var readFileFuncGo2 ReadFileFuncType2
	purego.RegisterFunc(&readFileFuncGo2, readFileFunc)

	// 假设文件类型为1（数据文件）
	fileType := uintptr(1)
	retCode = readFileFuncGo2(handle, fileType, fileID, offset, unsafe.Pointer(&buffer[0]), uintptr(len(buffer)))
	fmt.Printf("  调用结果 (6参数): 返回码=0x%08X (%s)\n", retCode, getErrorDescription(retCode))

	if retCode == DONGLE_SUCCESS {
		return retCode, len(buffer), nil
	}

	return retCode, 0, fmt.Errorf(getErrorDescription(retCode))
}

// closeDevice 关闭设备
func closeDevice(closeFunc uintptr, handle DongleHandle) (uint32, error) {
	if handle == 0 {
		return DONGLE_SUCCESS, nil // 已经关闭
	}

	if closeFunc == 0 {
		return DONGLE_SUCCESS, nil // 没有 Close 函数
	}

	// 使用 purego 的正确方式调用函数
	// 定义函数原型
	type CloseFuncType func(handle DongleHandle) uint32

	// 将函数指针转换为可调用的函数
	var closeFuncGo CloseFuncType
	purego.RegisterFunc(&closeFuncGo, closeFunc)

	retCode := closeFuncGo(handle)
	fmt.Printf("  关闭设备返回码: 0x%08X (%s)\n", retCode, getErrorDescription(retCode))

	if retCode != DONGLE_SUCCESS {
		return retCode, fmt.Errorf(getErrorDescription(retCode))
	}

	return retCode, nil
}

// showDeviceInfo 显示设备信息
func showDeviceInfo(keyList []DongleInfo, count int) {
	if count == 0 {
		fmt.Println("未找到任何设备")
		return
	}

	for i := 0; i < count; i++ {
		fmt.Printf("====== 设备 %d ======\n", i)
		fmt.Printf("版本号: %04X\n", keyList[i].MVer)
		fmt.Printf("生产日期: ")
		showBinHex(keyList[i].MBirthDay[:])
		fmt.Printf("代理商ID: %08X\n", keyList[i].MAgent)
		fmt.Printf("产品ID: %08X\n", keyList[i].MPID)
		fmt.Printf("用户ID: %08X\n", keyList[i].MUserID)
		fmt.Printf("是否母锁: %08X\n", keyList[i].MIsMother)
		fmt.Printf("硬件ID: ")
		showBinHex(keyList[i].MHID[:])
	}

	fmt.Printf("找到的设备数量: %d\n", count)
}

// ============ 测试函数 ============

// runPlatformTest 运行平台测试
func runPlatformTest() {
	fmt.Println("=== 运行平台测试 ===")
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("系统架构: %s\n", runtime.GOARCH)
	fmt.Printf("编译器: %s\n", runtime.Compiler)
	fmt.Printf("Go版本: %s\n", runtime.Version())

	// 检查是否在Linux平台
	if runtime.GOOS != "linux" {
		fmt.Printf("错误: 此程序仅支持Linux平台，当前平台: %s\n", runtime.GOOS)
		return
	}

	// 测试库路径获取
	fmt.Println("\n=== 库路径测试 ===")
	libPath := getLibraryPath()
	fmt.Printf("库文件路径: %s\n", libPath)

	// 检查库文件是否存在
	if _, err := os.Stat(libPath); err != nil {
		fmt.Printf("库文件不存在: %v\n", err)
		fmt.Println("注意: 这可能是正常的，如果没有实际的库文件")
	} else {
		fmt.Println("库文件存在")
	}

	// 尝试加载库
	fmt.Println("\n=== 动态库加载测试 ===")
	handle, err := loadLibrary(libPath)
	if err != nil {
		fmt.Printf("动态库加载失败: %v\n", err)
		fmt.Println("注意: 这可能是正常的，如果没有实际的库文件")
	} else {
		defer purego.Dlclose(handle)
		fmt.Println("动态库加载成功")

		// 尝试获取函数符号
		symbols := []string{
			FUNC_ENUM,
			FUNC_OPEN,
			FUNC_READFILE,
			FUNC_CLOSE,
		}

		for _, sym := range symbols {
			if _, err := getProcAddress(handle, sym); err != nil {
				fmt.Printf("符号 '%s' 获取失败: %v\n", sym, err)
			} else {
				fmt.Printf("符号 '%s' 获取成功\n", sym)
			}
		}
	}

	fmt.Println("\n=== 平台测试完成 ===")
}

// runDeviceTest 运行设备测试
func runDeviceTest() {
	fmt.Println("=== Rockey-ARM 设备测试 ===")
	fmt.Printf("操作系统: %s, 架构: %s\n", runtime.GOOS, runtime.GOARCH)

	// 检查是否在Linux平台
	if runtime.GOOS != "linux" {
		fmt.Printf("错误: 此程序仅支持Linux平台，当前平台: %s\n", runtime.GOOS)
		return
	}

	// 显示库文件路径
	libPath := getLibraryPath()
	fmt.Printf("库文件路径: %s\n", libPath)

	// 检查当前用户权限
	fmt.Printf("当前用户: ")
	cmd := exec.Command("whoami")
	if output, err := cmd.Output(); err == nil {
		fmt.Printf("%s", output)
	} else {
		fmt.Printf("未知 (错误: %v)\n", err)
	}

	// 加载库
	fmt.Println("\n加载动态库...")
	handle, err := loadLibrary(libPath)
	if err != nil {
		fmt.Printf("加载库失败: %v\n", err)
		return
	}
	defer func() {
		fmt.Println("关闭动态库...")
		purego.Dlclose(handle)
	}()

	// 获取函数地址
	fmt.Println("\n获取函数地址...")
	enumFunc, err := getProcAddress(handle, FUNC_ENUM)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v\n", FUNC_ENUM, err)
		return
	}

	openFunc, err := getProcAddress(handle, FUNC_OPEN)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v\n", FUNC_OPEN, err)
		return
	}

	readFileFunc, err := getProcAddress(handle, FUNC_READFILE)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v\n", FUNC_READFILE, err)
		return
	}

	closeFunc, err := getProcAddress(handle, FUNC_CLOSE)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v (可能是可选的)\n", FUNC_CLOSE, err)
		closeFunc = 0
	} else {
		fmt.Printf("获取 %s 成功\n", FUNC_CLOSE)
	}

	// 1. 枚举设备
	fmt.Println("\n1. 枚举设备...")
	keyList, count, retCode, err := enumDevices(enumFunc)
	if err != nil {
		fmt.Printf("设备枚举失败，错误码: %08X - %s\n", retCode, getErrorDescription(retCode))
		if retCode == DONGLE_NOT_FOUND {
			fmt.Println("提示: 请确保 Rockey-ARM 加密狗已正确连接到计算机")
		} else if retCode == DONGLE_UNKNOWN_ERROR {
			fmt.Println("提示: 未知错误，可能是:")
			fmt.Println("  1. 动态库版本不兼容")
			fmt.Println("  2. 函数调用参数不正确")
			fmt.Println("  3. 系统权限不足")
			fmt.Println("  4. 设备驱动程序未安装")
		}
		return
	}

	if count == 0 {
		fmt.Println("未找到任何 Rockey-ARM 设备")
		fmt.Println("可能的原因:")
		fmt.Println("  1. 加密狗未连接")
		fmt.Println("  2. 设备驱动程序未安装")
		fmt.Println("  3. 用户权限不足（尝试使用 sudo）")
		fmt.Println("  4. 设备被其他程序占用")
		return
	}

	// 显示设备信息
	showDeviceInfo(keyList, count)

	// 2. 打开第一个设备
	fmt.Println("\n2. 打开设备...")
	deviceHandle, retCode, err := openDevice(openFunc, 0)
	if err != nil {
		fmt.Printf("打开设备失败，错误码: %08X - %s\n", retCode, getErrorDescription(retCode))
		return
	}
	defer closeDevice(closeFunc, deviceHandle)

	// 3. 读取文件
	fmt.Println("\n3. 读取文件...")
	buffer := make([]byte, TEST_BUFFER_SIZE)
	retCode, dataSize, err := readFile(readFileFunc, deviceHandle, TEST_FILE_ID, TEST_OFFSET, buffer)
	if err != nil {
		fmt.Printf("读取文件失败，错误码: %08X - %s\n", retCode, getErrorDescription(retCode))
		return
	}

	if retCode == DONGLE_SUCCESS {
		fmt.Printf("成功读取 %d 字节数据\n", dataSize)

		// 显示前 64 字节
		if dataSize > 0 && len(buffer) > 0 {
			displaySize := dataSize
			if displaySize > 64 {
				displaySize = 64
			}
			fmt.Println("文件内容（前", displaySize, "字节）：")
			showBinHex(buffer[:displaySize])

			// 显示十六进制字符串
			fmt.Println("十六进制字符串：")
			fmt.Println(hex.EncodeToString(buffer[:displaySize]))
		}
	}

	fmt.Println("\n=== 设备测试完成 ===")
}

// runReadFileTest 运行读取文件测试
func runReadFileTest() {
	fmt.Println("=== Rockey-ARM 读取文件参数测试 ===")
	fmt.Printf("操作系统: %s, 架构: %s\n", runtime.GOOS, runtime.GOARCH)

	// 检查是否在Linux平台
	if runtime.GOOS != "linux" {
		fmt.Printf("错误: 此程序仅支持Linux平台，当前平台: %s\n", runtime.GOOS)
		return
	}

	// 显示库文件路径
	libPath := getLibraryPath()
	fmt.Printf("库文件路径: %s\n", libPath)

	// 加载库
	handle, err := loadLibrary(libPath)
	if err != nil {
		fmt.Printf("加载库失败: %v\n", err)
		return
	}
	defer purego.Dlclose(handle)

	// 获取函数地址
	enumFunc, err := getProcAddress(handle, FUNC_ENUM)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v\n", FUNC_ENUM, err)
		return
	}

	openFunc, err := getProcAddress(handle, FUNC_OPEN)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v\n", FUNC_OPEN, err)
		return
	}

	readFileFunc, err := getProcAddress(handle, FUNC_READFILE)
	if err != nil {
		fmt.Printf("获取 %s 失败: %v\n", FUNC_READFILE, err)
		return
	}

	closeFunc, _ := getProcAddress(handle, FUNC_CLOSE)

	// 枚举设备
	fmt.Println("\n1. 枚举设备...")
	_, count, retCode, err := enumDevices(enumFunc)
	if err != nil {
		fmt.Printf("设备枚举失败，错误码: %08X - %s\n", retCode, getErrorDescription(retCode))
		if retCode == DONGLE_NOT_FOUND {
			fmt.Println("提示: 请确保 Rockey-ARM 加密狗已正确连接到计算机")
		}
		return
	}

	if count == 0 {
		fmt.Println("未找到任何 Rockey-ARM 设备")
		return
	}

	// 打开第一个设备
	fmt.Println("\n2. 打开设备...")
	deviceHandle, retCode, err := openDevice(openFunc, 0)
	if err != nil {
		fmt.Printf("打开设备失败，错误码: %08X - %s\n", retCode, getErrorDescription(retCode))
		return
	}
	defer closeDevice(closeFunc, deviceHandle)

	// 测试不同的参数组合
	fmt.Println("\n3. 测试不同的读取文件参数组合...")
	testCases := []struct {
		Name   string
		FileID uintptr
		Offset uintptr
	}{
		{"文件ID=0x0000-偏移=0", 0x0000, 0},
		{"文件ID=0x0001-偏移=0", 0x0001, 0},
		{"文件ID=0x0000-偏移=100", 0x0000, 100},
		{"文件ID=0x0001-偏移=100", 0x0001, 100},
	}

	bufferSize := TEST_BUFFER_SIZE

	for i, tc := range testCases {
		fmt.Printf("测试 %d/%d: %s\n", i+1, len(testCases), tc.Name)

		buffer := make([]byte, bufferSize)
		retCode, dataSize, err := readFile(readFileFunc, deviceHandle, tc.FileID, tc.Offset, buffer)

		success := (retCode == DONGLE_SUCCESS)
		fmt.Printf("  结果: %s (错误码: %08X)\n",
			map[bool]string{true: "成功", false: "失败"}[success],
			retCode)

		if err != nil {
			fmt.Printf("  错误: %v\n", err)
		}

		if success && dataSize > 0 {
			fmt.Printf("  读取 %d 字节数据\n", dataSize)

			// 显示前 32 字节
			displaySize := dataSize
			if displaySize > 32 {
				displaySize = 32
			}
			fmt.Println("  数据（前", displaySize, "字节）：")
			showBinHex(buffer[:displaySize])
		}

		// 成功则停止测试
		if success {
			break
		}
	}

	fmt.Println("\n=== 读取文件测试完成 ===")
}

// runDiagnose 运行详细诊断
func runDiagnose() {
	fmt.Println("=== Rockey-ARM 详细诊断模式 ===")
	fmt.Printf("操作系统: %s, 架构: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("编译器: %s\n", runtime.Compiler)

	// 检查是否在Linux平台
	if runtime.GOOS != "linux" {
		fmt.Printf("错误: 此程序仅支持Linux平台，当前平台: %s\n", runtime.GOOS)
		return
	}

	// 1. 系统信息
	fmt.Println("\n1. 系统信息检查:")
	fmt.Printf("   当前用户: ")
	cmd := exec.Command("whoami")
	if output, err := cmd.Output(); err == nil {
		fmt.Printf("%s", output)
	} else {
		fmt.Printf("未知 (错误: %v)\n", err)
	}

	// 检查用户组
	fmt.Printf("   用户组: ")
	cmd = exec.Command("groups")
	if output, err := cmd.Output(); err == nil {
		groups := string(output)
		// 检查是否有USB相关权限
		if contains(groups, "plugdev") || contains(groups, "usb") || contains(groups, "dialout") {
			fmt.Printf("%s (包含USB权限组)\n", groups)
		} else {
			fmt.Printf("%s (可能缺少USB权限)\n", groups)
		}
	} else {
		fmt.Printf("未知 (错误: %v)\n", err)
	}

	// 2. 库文件检查
	fmt.Println("\n2. 库文件检查:")
	libPath := getLibraryPath()
	fmt.Printf("   库文件路径: %s\n", libPath)

	// 检查文件是否存在
	fileInfo, err := os.Stat(libPath)
	if err != nil {
		fmt.Printf("   ✗ 库文件不存在: %v\n", err)
		fmt.Println("   请确保库文件位于以下位置之一:")
		fmt.Println("     - ./lib/linux/arm64/libRockeyARM.so.0.3 (ARM64)")
		fmt.Println("     - ./lib/linux/loong64/libRockeyARM.so.0.3 (Loong64)")
		fmt.Println("     - ./lib/linux/libRockeyARM.so (x86_64)")
		return
	}

	fmt.Printf("   ✓ 库文件存在\n")
	fmt.Printf("     文件大小: %d 字节\n", fileInfo.Size())
	fmt.Printf("     文件权限: %v\n", fileInfo.Mode())
	fmt.Printf("     修改时间: %v\n", fileInfo.ModTime())

	// 3. 动态库加载测试
	fmt.Println("\n3. 动态库加载测试:")
	handle, err := loadLibrary(libPath)
	if err != nil {
		fmt.Printf("   ✗ 动态库加载失败: %v\n", err)
		fmt.Println("   可能的原因:")
		fmt.Println("     - 库文件损坏")
		fmt.Println("     - 缺少依赖库")
		fmt.Println("     - 架构不匹配")
		fmt.Println("     - 文件权限问题")
		return
	}
	defer purego.Dlclose(handle)
	fmt.Printf("   ✓ 动态库加载成功\n")
	fmt.Printf("     库句柄: 0x%x\n", handle)

	// 4. 函数符号检查
	fmt.Println("\n4. 函数符号检查:")
	symbols := []string{
		"Dongle_Enum",
		"Dongle_Open",
		"Dongle_ReadFile",
		"Dongle_Close",
		"Dongle_Write",
		"Dongle_Seed",
		"Dongle_Generate",
		"Dongle_WriteFile",
	}

	for _, sym := range symbols {
		addr, err := getProcAddress(handle, sym)
		if err != nil {
			fmt.Printf("   ✗ %s: 未找到 (%v)\n", sym, err)
		} else {
			fmt.Printf("   ✓ %s: 找到 (地址: 0x%x)\n", sym, addr)
		}
	}

	// 5. 设备文件检查
	fmt.Println("\n5. 设备文件检查:")
	fmt.Println("   检查USB设备:")
	cmd = exec.Command("lsusb")
	if output, err := cmd.Output(); err == nil {
		lsusbOutput := string(output)
		if len(lsusbOutput) > 0 {
			fmt.Printf("   USB设备列表:\n%s", lsusbOutput)

			// 检查是否有类似加密狗的设备
			if contains(lsusbOutput, "Rockey") || contains(lsusbOutput, "Feitian") || contains(lsusbOutput, "HID") {
				fmt.Println("   ✓ 发现可能的加密狗设备")
			} else {
				fmt.Println("   ⚠ 未发现明显的加密狗设备")
			}
		} else {
			fmt.Println("   无USB设备")
		}
	} else {
		fmt.Printf("   无法执行lsusb: %v\n", err)
		fmt.Println("   尝试安装lsusb: sudo apt-get install usbutils")
	}

	// 检查设备文件权限
	fmt.Println("\n   检查设备文件权限:")
	devicePatterns := []string{
		"/dev/usb/hiddev*",
		"/dev/bus/usb/*/*",
		"/dev/hidraw*",
		"/dev/ttyUSB*",
		"/dev/ttyACM*",
	}

	foundDevices := false
	for _, pattern := range devicePatterns {
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			for _, device := range matches {
				if info, err := os.Stat(device); err == nil {
					fmt.Printf("     %s: 权限 %v\n", device, info.Mode())
					foundDevices = true
				}
			}
		}
	}

	if !foundDevices {
		fmt.Println("     未找到相关设备文件")
	}

	// 6. 内核模块检查
	fmt.Println("\n6. 内核模块检查:")
	cmd = exec.Command("lsmod")
	if output, err := cmd.Output(); err == nil {
		lsmodOutput := string(output)
		if contains(lsmodOutput, "usbhid") || contains(lsmodOutput, "hid") || contains(lsmodOutput, "usb") {
			fmt.Println("   ✓ USB/HID相关内核模块已加载")
		} else {
			fmt.Println("   ⚠ USB/HID相关内核模块可能未加载")
		}
	} else {
		fmt.Printf("   无法检查内核模块: %v\n", err)
	}

	// 7. 权限建议
	fmt.Println("\n7. 权限建议:")
	fmt.Println("   如果遇到权限问题，可以尝试:")
	fmt.Println("     - 使用sudo运行程序: sudo ./rockey-test -test")
	fmt.Println("     - 将用户添加到相关组: sudo usermod -a -G plugdev $USER")
	fmt.Println("     - 创建udev规则: sudo nano /etc/udev/rules.d/99-rockey.rules")
	fmt.Println("       添加: SUBSYSTEM==\"usb\", ATTR{idVendor}==\"****\", ATTR{idProduct}==\"****\", MODE=\"0666\"")
	fmt.Println("     - 重新登录使组更改生效")

	// 8. 测试建议
	fmt.Println("\n8. 测试建议:")
	fmt.Println("   如果诊断通过但设备测试失败，请尝试:")
	fmt.Println("     - 重新插拔加密狗")
	fmt.Println("     - 检查加密狗指示灯")
	fmt.Println("     - 在其他电脑上测试加密狗")
	fmt.Println("     - 检查库文件版本是否与硬件匹配")
	fmt.Println("     - 查看系统日志: dmesg | tail -20")

	fmt.Println("\n=== 诊断完成 ===")
}

// printHelp 显示帮助信息
func printHelp() {
	fmt.Println("Rockey-ARM 测试程序 (Linux版)")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  rockey_test [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -test          运行设备测试（测试加密狗功能）")
	fmt.Println("  -platform      运行平台测试（测试Linux平台兼容性）")
	fmt.Println("  -read-test     运行读取文件参数测试（测试不同参数组合）")
	fmt.Println("  -diagnose      运行详细诊断模式")
	fmt.Println("  -h, -help     显示帮助信息")
	fmt.Println()
	fmt.Println("描述:")
	fmt.Println("  这是一个Linux平台的Rockey-ARM加密狗测试程序。")
	fmt.Println("  使用purego纯Go实现动态库加载，无需CGO。")
	fmt.Println()
	fmt.Println("支持的架构:")
	fmt.Println("  - x86_64/amd64")
	fmt.Println("  - arm64/aarch64")
	fmt.Println("  - loong64")
	fmt.Println()
	fmt.Println("构建:")
	fmt.Println("  go build -o rockey-test rockey_test.go")
	fmt.Println("  CGO_ENABLED=0 go build -o rockey-test-static rockey_test.go  # 纯Go静态构建")
}

// ============ 主函数 ============

func main() {
	flag.Parse()

	// 显示帮助信息
	if *help || *helpLong {
		printHelp()
		return
	}

	// 运行详细诊断
	if *diagnoseMode {
		runDiagnose()
		return
	}

	// 运行平台测试
	if *platformTest {
		runPlatformTest()
		return
	}

	// 运行读取文件测试
	if *readTest {
		runReadFileTest()
		return
	}

	// 运行设备测试
	if *testMode {
		runDeviceTest()
		return
	}

	// 默认显示帮助
	printHelp()
}
