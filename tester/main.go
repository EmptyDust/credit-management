package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("开始运行用户管理服务测试...")

	// 运行测试
	cmd := exec.Command("go", "test", "-v", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("测试执行失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("测试执行完成！")
}
