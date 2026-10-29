# Main Profile 与共享 AGENTS.md 编辑

## 目标与交互

Main 保留一个统一的 Profile 页面，按指令（AGENTS.md）、模型、技能和 MCP 分区。打开后自动读取；阅读态展示整份配置，编辑态统一保存或取消。不把 Main 的配置混入 Worker 技能空间或 Worker MCP。

共享 AGENTS.md 默认展示完整的 Markdown 文档，编辑时切换为大文本编辑区；保存后回到文档阅读态。移除原来的重复“最新已保存指令”、默认小文本框和纵向堆叠按钮。刷新、编辑、保存、取消集中在工具栏。切换资源栏目保留未提交草稿，切换 Runtime 则按 epoch/scope 隔离状态。

## Runtime 契约适配

对照本地 `agentkit-mpa-agent/app/schemas/main_profile.py`、`app/api/v1/main_profile.py` 和 `app/services/main_profile.py`：

- 完整读取和保存 `agentsMd`、`model`、`skills`、`mcpSources`、`mcpServers`；已有不含 MCP 字段的 Profile 保持兼容。未知新字段仍拒绝读取，防止全量替换时丢失配置。
- 模型支持继承或固定模型 ID。模型需要由 Main 部署的模型端点支持，不把 Worker 模型目录误当成 Main 的保证。
- 技能是固定版本的说明正文，不支持 `latest` 或浮动版本，不安装脚本；指令与技能合计最多 40,000 字符。
- MCP 支持目录资源 ID 引用、HTTPS Streamable HTTP 直接连接及请求头。工具白名单保留 null（全部）、非空列表（指定）、空列表（禁用）三种语义。
- Runtime 返回的 MCP 请求头是 `**********`。编辑时以“已保存，请重新输入”提示，写入前拒绝脱敏占位符或加密密文；不把真实新值写入日志、文档或本地持久化。取消或保存成功后清除编辑态中的真实值。Runtime 的整份 PUT 契约要求重填保留的请求头；客户端不能承诺无感保留密钥。
- 保存使用原生持有的单次快照、`If-Match` 和幂等键，保留来源元数据。冲突或结果未知保留草稿、禁止直接重放；重新读取后保留已编辑字段并更新未编辑字段。仅重填相同凭据的服务端无操作响应可保留原版本。
- 未找到 Profile 时不自动创建。用户显式选择新建、编辑并保存后，原生端通过 Runtime 元数据解析实例身份，发送带 `If-None-Match: *` 的仅新建请求；412/409 不覆盖已存在配置。404 仍可能是旧 Runtime 不支持接口，创建入口不代表该 Runtime 已支持。

本次不修改 `agentkit-mpa-agent`，不发布 Runtime，也不写入真实云端配置。

## 验证

- 前端覆盖自动读取、完整保存、凭据重填、取消清除真实值、文档阅读/编辑、冲突恢复、编辑中的异步保存、显式创建及新 IPC 参数映射。
- Go 覆盖完整 Profile 往返、MCP 引用保留、工具范围三态、凭据与 URL 校验、创建 CAS、凭据无操作版本，以及原有单次快照、作用域和未知结果保护。
- 视觉与交互通过本地模拟 IPC 的真实组件验证：Main 阅读/编辑/保存、AGENTS.md 阅读/编辑/保存、深色和窄窗口布局。未进行真实 Runtime 的保存联调。
