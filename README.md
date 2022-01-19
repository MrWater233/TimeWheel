# TimeWheel
Golang实现的简单时间轮算法
## 原理
https://cloud.tencent.com/developer/article/1815722
## 安装
```shell
go get -u github.com/MrWater233/timewheel
```
## 使用
```go
package main

import (
	"github.com/MrWater233/timewheel"
	"fmt"
	"time"
)

func main() {
	// 创建时间轮
	// 参数一：转动触发时间
	// 参数二：时间轮槽数
	// 参数三：回调函数
	tw := timewheel.New(time.Second, 3600, func(i interface{}) {
		// 回调事件
		fmt.Println(time.Now().Format("2006-01-02 03:04:05"))
	})

	// 启动时间轮
	tw.Start()
	fmt.Println(time.Now().Format("2006-01-02 03:04:05"))

	// 添加定时任务
	// 参数一：延迟时间
	// 参数二：标识，用于删除定时任务
	// 参数三：传给回调函数的参数
	tw.AddTask(5*time.Second, "key1", "hello")

	// 删除定时任务，参数为定时任务标识
	tw.RemoveTask("key1")

	// 停止时间轮
	tw.Stop()

	select {}
}

```
