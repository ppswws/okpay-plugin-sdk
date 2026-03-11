# payment/plugin 独立发布说明

当前已将 `payment/plugin` 设为独立 Go module：
- `go.mod`：`module okpay/payment/plugin`

如果要发布到 GitHub 并供外部插件直接引用，建议：

1. 修改 `payment/plugin/go.mod` 的 `module` 为你的仓库路径  
示例：`module github.com/<owner>/<repo>/payment/plugin`

2. 批量替换 `payment/plugin` 内部 import 路径  
把 `okpay/payment/plugin/...` 改为新的 module 前缀。

3. 在独立仓库（或子目录）执行：
- `go mod tidy`
- `go build ./...`

4. 外部插件改为直接依赖新模块路径（仅依赖 `payment/plugin`）。
