# 贡献指南

欢迎通过 Issue 说明问题，或提交范围清晰的小 PR。先阅读 [README](README.md)、[开发指南](docs/development.md)和[源码导航](src/SOURCE_STRUCTURE.md)。

## 准备环境

需要 Go 1.25.0 起、Make、CGO 和 C 编译器；非 macOS 还需满足 Rust 静态库的链接条件。所有命令从仓库根目录执行：

```bash
go mod download
make init-config
make config-test
```

`init-config` 默认目标为 `local/server_local_dev.yaml`，可用 `CONFIG` 指定其他路径，不覆盖已有文件。先核对本地配置，再按需 `make api`；配置详情见[配置说明](docs/configuration.md)。启动会尝试迁移数据库，不要连接生产库做开发验证。

## 保持 PR 小而明确

- 一个 PR 解决一个问题，说明触发条件、改前/改后行为和影响范围。大范围接口或数据结构调整宜先讨论。
- 文档、移动文件、同包拆分等“行为不变”改动应保留导出 API、包路径、YAML/JSON 键、默认值、路由、鉴权、迁移顺序、错误处理、任务触发条件和初始化副作用；不要顺便修复业务逻辑。
- 业务改动说明兼容性和数据影响，选择能验证问题的测试；不要把格式清理、依赖升级或第三方同步混入无关 PR。
- 新增配置、命令、路由或脚本时同步更新对应文档，示例必须对应真实代码。不要承诺未验证的能力或测试结果。

## 依赖、第三方代码与本地文件

Go 依赖以 `go.mod`、`go.sum` 为准。`vendor/` 被 `.gitignore` 忽略，工作区里存在它不代表公开克隆包含它，也不要把本地 vendor 整包加入 PR。已有 vendor 时，Go 命令可能使用该目录；需要按模块解析时显式使用 `-mod=mod`，并说明验证方式。

依赖调整应有实际用途，检查 `go.mod` / `go.sum` 差异；不要为文档修改执行 `go mod tidy` 或无关升级。

以下路径含第三方代码或数据，修改前核对来源、版本与许可，保留署名：

- `src/robot/qq_bot/Better-Bot-Go/`、`src/robot/qq_bot/Bot-Client-Go/`。
- `build/Alg-Trainers/`、`build/Algs/`，导入规则见 [scripts/README.md](scripts/README.md)。提交前核对 `git ls-files` 及新增文件，不能仅凭本地目录内容判断公开克隆内容。
- `src/internel/scramble/librust_scramble.a` 及头文件、字体等原生或资源文件。不要在没有来源依据时替换二进制或编造上游链接。

`local/`、`test/`、日志、缓存以及其他忽略项属于本地工作数据。不要提交真实 DSN、密码、OAuth/机器人令牌、私钥、用户数据或抓取快照；提交前检查暂存差异，而非依赖忽略规则保密。

## 验证要求

仓库没有“任意环境可安全运行全量测试”的保证。先读目标包及其导入链中的 `init()`、测试夹具和文件/网络操作，再选择验证范围。`go test -run` 仍会执行包初始化；禁用 `-j` / `-w` 也不能屏蔽导入阶段的副作用。

可按改动选择以下验证，具体条件见[开发指南](docs/development.md)：

```bash
# 基础配置包，不加载 svc
make config-test
# 群比赛逻辑：数据库测试需先配置 CUBING_TEST_MYSQL_DSN，未配置时跳过
go test -mod=mod ./src/internel/groupcomp
# 根入口构建：不执行应用 init，但仍需编译资源和原生依赖
make build
# 文档和空白差异检查
git diff --check
```

涉及真实数据库、远程同步、邮件、机器人或生成文件的测试必须显式选取包与测试名，使用隔离环境，并先确认会访问或改写的资源。不要将 `go test ./...` 或 `go test ./... -run '^$'` 当作默认无副作用检查。

PR 描述至少包含：

- 修改目的、范围与行为是否变化。
- 实际运行的命令、结果及平台/原生依赖条件。
- 未执行的检查及原因；只读核对、构建成功与测试通过应分开表述。
- 数据、配置、接口或第三方资源变化的迁移/复现说明（如适用）。

文档专属修改可仅做源码、路径、命令与差异核对，明确说明未启动应用或执行测试。不要为了填验证清单运行有外部副作用的测试。
