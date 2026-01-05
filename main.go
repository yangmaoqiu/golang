// test_minimal.go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("=== 龙芯最小化测试 ===")
	fmt.Printf("GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())

	// 测试内存分配
	slice := make([]byte, 100)
	fmt.Printf("Slice allocated: %d bytes\n", len(slice))

	// 测试goroutine
	done := make(chan bool)
	go func() {
		fmt.Println("Goroutine: Hello from another goroutine!")
		done <- true
	}()
	<-done

	fmt.Println("=== 测试通过 ===")
}
