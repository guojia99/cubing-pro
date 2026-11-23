# 手动实验工具

## CPU 负载实验服务

[`loadtest/main.go`](loadtest/main.go) 是独立的 CPU 负载实验服务，使用项目已有的 Gin 依赖。仅在明确需要负载实验时，从**仓库根目录**手动启动：

```bash
go run ./tools/loadtest
```

程序监听 `:8080`，即所有网卡的 8080 端口；以下三个 HTTP 路径均为 GET，且没有鉴权：

| 路径 | 行为 |
| --- | --- |
| `/start-high-load` | 按 `runtime.NumCPU()` 启动工作协程，在全部 CPU 核心上执行持续忙循环 |
| `/stop-high-load` | 关闭停止通道并等待负载协程退出 |
| `/status` | 返回 `highLoadActive` 状态 |

现有并发状态未同步：HTTP 处理函数共享 `highLoadActive`、任务指针和 WaitGroup，并发请求可能触发数据竞争或重复关闭通道等问题。实验前应控制网络访问，并串行发出控制请求。

**禁止将该程序作为默认测试或生产服务运行。** 它会主动制造全 CPU 负载，可能影响同机服务；不应纳入默认启动流程、测试流程或生产部署。移动文件不改变程序内容或当前工作目录（cwd）语义。
