# Runtime 使用侧适配验收

Change ID：desktop-runtime-usage。日期：2026-10-27。
状态：本地实现与验证完成；无 commit、push、部署或线上准出。
契约：`2026-10-27-runtime-usage-spec.md`。

## 实现与范围

| 验收         | 可观察结果                                                                                   | 结果                      |
| ------------ | -------------------------------------------------------------------------------------------- | ------------------------- |
| AC1 消息渠道 | 复用飞书、钉钉、企微原生绑定、权限及 Bot 配置；现有渠道回归通过                              | 本地 pass；真实写入未执行 |
| AC2 Skill    | 发现当前 Runtime 挂载空间，签名读取目录，分页展示名称/可选描述；选择填入会话草稿             | pass                      |
| AC3 MCP      | 展示当前会话的服务器和工具明细，连接现有模型/MCP 版本化配置入口                              | pass                      |
| AC4 指令     | 原生 AGENTS.md 编辑、revision、冲突/未知写入恢复；兼容已有 Main Profile，保留模型/固定 Skill | pass                      |
| AC5 A2A      | 空间/成员分页、绑定与解绑、CAS 别名；要求已保存的唯一别名，地址位于草稿开头且保留旧正文      | pass                      |
| AC6 身份隔离 | Studio key 由 Go 固定路由注入；重定向、过期连接/页面/会话结果受控；跨账户清空草稿            | pass                      |
| AC7 恢复     | 失败写入保留草稿并要求重读；不自动重发；发送回执只清理已发送文本                             | pass                      |

执行继续由 Runtime 承担。没有迁移资源创建、Skill 发布或 A2A 注册流程。
当前 Runtime 的成员目录不提供 Agent Card，页面如实显示不可用；没有从显示名称猜测寻址。
SKILL.md 内容编辑/发布继续由 Studio 承担。Main Profile 未配置/旧接口时不自动初始化。

## 验证证据

| 命令或环境                                                                                                                       | 结果                                                                         |
| -------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `GOTOOLCHAIN=go1.27.3 go test ./...`（desktop）                                                                                  | 全包通过                                                                     |
| `GOTOOLCHAIN=go1.27.3 go test -race ./internal/usage ./internal/connection ./internal/cloud ./internal/nativeui ./internal/host` | 受影响包通过                                                                 |
| `GOTOOLCHAIN=go1.27.3 go vet ./...`                                                                                              | 通过                                                                         |
| `LANG=en_US.UTF-8 npm test -- --maxWorkers=2`                                                                                    | 80 文件、428 项通过                                                          |
| `npm run typecheck`、`npm run lint`                                                                                              | 通过                                                                         |
| `npm run desktop:generate`                                                                                                       | 官方绑定与本地化生成通过                                                     |
| `npm run build:desktop:frontend`                                                                                                 | 通过；保留既有 Vite 配置及 bundle 大小提示                                   |
| `GOWORK=off GOTOOLCHAIN=go1.27.3 CGO_ENABLED=0 go build ./...`                                                                   | 通过                                                                         |
| 参考 Runtime 现有测试与实际 FastAPI 路由                                                                                         | 56 项测试、27 个响应检查通过                                                 |
| Go usage/connection 直连隔离 Uvicorn Runtime                                                                                     | 5 类契约、16 个 HTTP 请求通过                                                |
| 原生 `ui.NewTester`                                                                                                              | 表单、键盘、禁用状态、分页、草稿、epoch 与失败恢复通过；为 headless 控件验证 |
| 两轮审查与修复台账校验                                                                                                           | 5 个本轮修复项已关闭；验证器通过                                             |

未设英语 locale 的首次 Node 全量运行出现 8 项既有英文选择器失败；
新发送回执回归另发现 1 项真实问题，修复后全量通过。原生未知错误测试发现
错误文字被通用 accessibility label 替代，已改成独立错误容器并通过中英文验证。
失败记录保留，不将重跑通过改写为首次通过。

独立实现与审查使用主会话模型配置；共享文件单写。原生额外独立审查因中断停止，
由主 Agent 两轮自审与针对性测试补齐，未冒称完成独立审查。
原 13 个脏文件的业务身份诊断和本地化等已有改动保留；基线补丁在工作区外。

## 未执行与交付边界

真实 Bot 写入、云端 Skill/A2A 权限、实际 A2A 分发、模型/MCP 执行、PostgreSQL
指令并发、物理桌面操作及最终 app 打包验证未执行。本地 Runtime 使用临时 SQLite、
mock principal 与云目录 transport；不构成线上可用性或发布许可。
本次代码没有要求修改或部署 agentkit-mpa-agent/veadk-python。
详细真实与 mock 层见 `2026-10-27-usage-runtime-evidence.md`；
独立审查及修复见 `2026-10-27-usage-backend-review.md`。
