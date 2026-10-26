# Runtime Key 创建 Session Hotfix

## 1. 文档状态

- 状态：已接受
- 日期：2026-10-26
- 类型：Bug 修复，L3（身份与授权边界）
- Change ID：`runtime-key-session-create-hotfix`
- 关联：用户报告“设置 Runtime 后仍无法创建会话”；补充 [现有 OpenAPI 客户端合同](2026-10-25-existing-api-client-spec.md) 与 [原生内容实施契约](2026-10-26-native-content-spec.md)，不替代其中其它决策。
- 基线：`af85b28c021e162946a236d4a63d0f1c3dfbe839`（v0.1.0）。

## 2. 问题、证据与根因

结论：Runtime 已连接且 Runtime Key 已由宿主持有时，AgentKit key-auth 已提供可创建 Session 的服务端主体；桌面额外要求显式业务身份，造成客户端本地误拦截，并把身份缺失错误映射成“未连接 Runtime”。

| ID | 类型 | 结论/假设 | 精确证据 | 设计影响 | 验证/关闭方式 |
|---|---|---|---|---|---|
| FACT-1 | 已验证事实 | 用户界面同时显示 Connected Runtime 和 “Choose and connect a Runtime first” | 用户 2026-10-26 截图；基线 `af85b28:desktop/internal/nativeui/settings.go:220-221` 与 `af85b28:desktop/internal/nativeui/prompts.go:145-148` | 显示状态与创建前置条件矛盾 | 回归测试重建同一状态 |
| FACT-2 | 已验证事实 | 创建助手把 `!Connected` 与 `!HasBusinessIdentity` 合并为同一 `notConnected` | 基线 `af85b28:desktop/internal/nativeui/prompts.go:145-148` | 无法区分连接、凭据和可选身份 | 拆分创建准入判断 |
| FACT-3 | 已验证事实 | 已绑定连接会将 Runtime Key 作为 Bearer；只有显式身份存在时才增加 `X-User-Id` 或 `X-Jwt-Token` | `desktop/internal/connection/connection.go:96-105,107-155` | Runtime Key 是基础授权，显式业务身份是可选覆盖 | HTTP 合同测试 |
| FACT-4 | 已验证事实 | AgentKit key-auth 在没有 `x-user-id` 时使用固定主体 `agentkit-key-auth`，并授予 `session:create/read/update/delete` | 只读参考 `agentkit-mpa-agent/app/api/auth.py:24-38,182-217,336-348` | Runtime-key-only 请求具备服务端定义的 Session 身份与权限 | 本地 HTTP 合同 + 目标 Runtime 实测 |
| FACT-5 | 已验证事实 | Runtime 绑定拒绝缺少 Runtime Key 的目标 | `desktop/internal/ipc/services.go:77-90` 的 `bindRuntime` | 生产设置路径成功连接后必然具有 Runtime Key | 无 Key 负控保留失败 |
| FACT-6 | 已验证事实 | v0.1.0 默认 Node/Go/提交门禁通过但未覆盖 Runtime-key-only 创建助手 | v0.1.0 `build-verification.json`；当前 `desktop/internal/nativeui/prompts_test.go` 无该场景 | 确认门禁逃逸 `GG-1` | 把回归加入默认 `go test ./...` |

### 2.1 故障链路

1. 用户在原生 Settings 选择 Runtime；`Platform.BindRuntime` 成功绑定 endpoint 与 Runtime Key。
2. `Connection.Status` 正确返回 `Connected=true`、`HasCredential=true`、`HasBusinessIdentity=false`。
3. Settings 正确显示 Connected Runtime。
4. 用户创建助手时，`createPromptItem` 把第 2 步的合法状态判为 `notConnected`，请求没有到达 Session API。
5. 全局安全文案将 `notConnected` 显示为“Choose and connect a Runtime first”，掩盖真实判定条件。

根因是客户端准入条件复制并强化了服务端身份策略，而不是连接事件迟到、Runtime 切换丢失或服务端拒绝。

### 2.2 GateGap

`GG-1` 状态为 `planned`：v0.1.0 的默认测试覆盖显式业务身份及连接切换，但缺少“Connected + Runtime Key + 无显式业务身份”的负控/正向组合。稳定承载检查为 `TestAssistantCreationAllowsRuntimeKeyIdentity`，缺陷态负控是旧判定返回 `notConnected`，修复态要求准入；同一测试还要求 disconnected 返回 `notConnected`、无 Runtime Key 返回 `authRequired`。该检查进入默认 `go test ./...`，关闭证据必须包含先红后绿结果。

## 3. 目标、非目标与范围

目标：

1. 已通过 AgentKit OpenAPI 绑定、持有 Runtime Key 的 Runtime 可以直接创建独立 Session。
2. 显式业务身份继续作为可选的 Session 主体覆盖，不因 hotfix 被删除或自动伪造。
3. 设置页和创建流程对 Runtime Key 默认主体的说明一致，不再提示用户重复连接。

非目标：不修改 Runtime/Worker 源码、镜像或 OpenAPI；不把 IAM 用户、本地 Space、昵称或账户 ID 自动转换为业务 `user_id`；不放宽 channel/schedule 的独立身份要求；不改变 Session 幂等、回执、scope、epoch 或清理逻辑。

范围内：原生助手创建准入、设置页连接说明、WebView 连接状态投影、对应 Go/Node 回归与本机应用复验。范围外：知识、渠道、文件对话框和镜像采纳的既有准出缺口。

## 4. 需求与场景

| ID | 来源 | 优先级 | 触发/适用范围 | 必须行为或约束 | 可观察结果/度量 |
|---|---|---|---|---|---|
| FR-1 | 用户原始请求、FACT-1 至 FACT-5 | 必须 | Connected 且 HasCredential，未配置显式业务身份 | 允许进入既有 Session Read/Create/Rename 流程 | 不产生本地 `notConnected`；创建结果出现在 Session 列表 |
| FR-2 | FACT-2 | 必须 | 未连接 Runtime | 在任何远端写入前拒绝 | 显示连接 Runtime 的提示，Session API 调用数为 0 |
| FR-3 | FACT-3、FACT-5 | 必须 | Connected 但无 Runtime Key | 在任何远端写入前拒绝为认证缺失 | 显示 Authentication required，Session API 调用数为 0 |
| FR-4 | FACT-1、FACT-4 | 必须 | Runtime-key-only 连接展示 | 明确默认主体可直接会话、显式业务身份可选 | Settings 与 WebView 状态不再显示 Identity required |
| NFR-1 | 现有 OpenAPI 客户端合同 | 必须 | 所有上述路径 | 不把云身份或 UI 字段伪造成 `X-User-Id`，不泄露 Runtime Key | HTTP 头仍只由 Go connection 层生成，安全投影不含凭据 |
| NFR-2 | FACT-6 | 必须 | 默认本地 admission | GateGap 回归必须由 `go test ./...` 自动执行 | 缺陷判定复原时稳定失败，修复时通过 |

主场景：Given 用户已绑定 AgentKit Runtime 且未配置 Authentication，When 新建 assistant，Then 客户端使用既有 Runtime Key 创建并命名 Session。失败场景：未连接或凭据缺失时，客户端分别给出连接/认证错误且不提交远端写入。

## 5. 方案与状态契约

| ID | 决策 | 依据 | 影响范围 | 重新评估条件 |
|---|---|---|---|---|
| DEC-1 | 助手创建前置条件改为 `Connected && HasCredential` | FACT-3 至 FACT-5 | 仅原生 assistant 创建 | AgentKit 撤销 key-auth 默认主体或控制面支持无 Key 连接 |
| DEC-2 | `HasBusinessIdentity` 只表达显式覆盖，不再表达会话可用性 | FACT-4 | 设置说明和前端状态文本 | 服务端要求所有 Session 必须携带显式主体 |
| DEC-3 | 不新增 header、身份持久化或自动映射 | NFR-1 | connection/ipc 保持不变 | 出现经审查的新身份协议 |

准入函数返回三态：未连接为 `notConnected`；无 Runtime Key 为 `authRequired`；其余返回 nil。通过后完全复用现有 original SessionID intent、404 后 Create、Rename、回读和 `sessionCreated` 通知。显式身份已配置时继续发送原 header；未配置时只发送 Runtime Key，由服务端解析默认主体。

设置页连接成功文案改为“可开始会话；显式业务身份为可选覆盖”。Runtime 页同时显示当前使用 Runtime key default 或 Explicit business identity。WebView 的 `ConnectionVM.status` 使用相同语义。无 schema、IPC 方法、持久化格式或远端 API 变化。

## 6. 边界、失败、安全与恢复

| 场景 | 处理 | 状态变化 | 信号 | 恢复 |
|---|---|---|---|---|
| Runtime 未连接 | 本地 `notConnected` | 不创建 intent/Session | 原生错误文本 | 返回 Settings 连接 Runtime |
| Runtime Key 缺失 | 本地 `authRequired` | 不创建 intent/Session | 原生认证错误 | 重新绑定合法 Runtime |
| Key 被服务端拒绝 | 保留服务端 401/403 安全投影 | intent 按原未知/失败规则保留 | HTTP 状态与安全错误码 | 重新绑定/刷新，不自动重发 |
| Runtime-key-only 创建响应丢失 | 沿用原 SessionID intent，先 Read 后决定是否 Create | 不生成第二 ID | 既有 unknown/recovery 信号 | 显式恢复原 intent |
| 显式身份随后配置 | identity 操作递增 epoch，旧目录失效 | 新 scope 重新读取 Sessions | connectionChanged | 不混用默认主体目录 |

安全不变量：Runtime Key 仍只在 Go 宿主持有；前端不获得 Key；客户端不构造默认主体字符串、不发送伪造 `X-User-Id`；服务端继续是主体与权限权威。滥用场景“用云账户 ID 代替业务用户”仍被非目标和现有表单显式输入边界阻止。

## 7. 备选、兼容、回滚与发布

备选 A：连接 Runtime 后自动填充云账户 ID。拒绝，因为云主体与业务主体是不同信任边界，会伪造 Session 所有权。备选 B：保留硬阻断，仅改错误文案。拒绝，因为它仍与 AgentKit key-auth 的服务端契约冲突，不能满足已连接后创建 Session 的用户结果。

hotfix 向后兼容：显式 JWT/test subject 的行为不变；本地状态文件无需迁移；旧版与新版可连接同一 Runtime，但看到的默认主体目录由服务端权威决定。回滚只需恢复客户端准入与文案，无远端 schema 或数据回滚；已创建 Session 仍由服务端保留。

发布前先跑局部测试、Go 全包、Node 相关测试、类型/lint/格式、绑定生成检查与 app 构建。本机安装只替换 OpenMPA app，不修改 Runtime。若 Runtime-key-only 实测返回 401/403，停止 hotfix 准出并保留安全错误，不改用伪造身份绕过。

## 8. 实施计划、验收与追踪

1. S1 / GG-1：加入准入回归，证明旧逻辑失败；安全停止点是仅测试与本文档。
2. S2：拆出纯准入判定，调整原生创建与设置/前端文案；完成证据是局部回归通过。
3. S3：运行默认门禁和只读参考契约复核；随后用目标 AgentKit Runtime 创建一个专用 Session 并按精确 ID 删除。
4. S4：重建本机 app，验证选择/恢复 Runtime 后可创建会话；不自动发布或更新 v0.1.0 标签。

| ID | Given | When | Then | 验证方法/命令 | 环境 |
|---|---|---|---|---|---|
| AC-1 | Connected=true、HasCredential=true、HasBusinessIdentity=false | 评估助手创建准入 | 返回允许 | `go test ./internal/nativeui -run TestAssistantCreationAllowsRuntimeKeyIdentity` | 本地 |
| AC-2 | 未连接或无 Runtime Key | 评估助手创建准入 | 分别返回 notConnected/authRequired | 同一表驱动回归 | 本地 |
| AC-3 | Runtime-key-only 目标连接 | 创建并命名专用 Session | API 成功且列表可见，精确删除回读 404 | 私有 live harness，不打印凭据 | 测试 Runtime |
| AC-4 | 显式业务身份已配置 | 创建/读取 Session | 行为与 hotfix 前一致 | 现有 native/session 回归 | 本地 + 既有 live 证据 |
| AC-5 | hotfix 源码 | 执行默认 admission 和构建 | 测试、类型、lint、格式、绑定与构建通过 | `go test ./...`、`npm test` 等仓库命令 | 本地 |
| AC-6 | 安装 hotfix app | 选择或恢复 Runtime 后新建 assistant | 不出现截图错误，Session 成功进入目录 | 本机原生实测 | macOS |

| 需求 | 方案位置/DEC | 验收标准 | 验证方法 | 覆盖状态 |
|---|---|---|---|---|
| FR-1 | 第 5 节 / DEC-1 | AC-1、AC-3、AC-6 | Go 回归 + live + 原生 | 已覆盖 |
| FR-2 | 第 5、6 节 | AC-2 | Go 表驱动测试 | 已覆盖 |
| FR-3 | 第 5、6 节 | AC-2 | Go 表驱动测试 | 已覆盖 |
| FR-4 | 第 5 节 / DEC-2 | AC-6 | 状态文本与原生实测 | 已覆盖 |
| NFR-1 | 第 5、6 节 / DEC-3 | AC-3、AC-4 | HTTP 头合同 + live | 已覆盖 |
| NFR-2 / GG-1 | 第 2.2、8 节 | AC-1、AC-2、AC-5 | 先红后绿 + 默认 Go admission | 已覆盖 |

## 9. 风险与待确认项

| ID | 触发条件 | 概率/影响 | 缓解措施 | 监测信号 | 兜底/回滚 |
|---|---|---|---|---|---|
| RISK-1 | 非 AgentKit、自托管且 JWT-disabled 的 Runtime 也被控制面返回为相同连接 | 低/高 | 生产路径仍要求 Runtime Key；实测目标 AgentKit Runtime；服务端 401/403 失败关闭 | authRequired/denied | 恢复显式身份准入，不伪造 user ID |
| RISK-2 | 多个桌面操作者共享 Runtime-key 默认主体 | 中/中 | UI 明确默认主体；需要隔离时配置显式业务身份 | Connection status 与 Session 目录 | 配置显式身份后 epoch/scope 切换 |
| RISK-3 | 设置文案修复但其它入口仍使用旧判定 | 低/中 | 准入集中为纯函数，原生入口统一调用 | 默认回归 | 回滚单一 helper 调用 |

无阻塞待确认项。若 AC-3 与服务端只读契约相矛盾，则将该结果作为新事实重新评审 DEC-1，不扩大身份或权限范围。
