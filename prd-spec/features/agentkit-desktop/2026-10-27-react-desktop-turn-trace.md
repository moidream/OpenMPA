# React Desktop Turn Trace 入口与可视化设计

## 1. 文档状态

| 字段 | 内容 |
|---|---|
| 标题 | React Desktop Turn Trace 入口与可视化设计 |
| 状态 | 评审中 |
| 创建日期 | 2026-10-27 |
| 变更类型 | 功能 |
| 关联任务 | 无 |
| 关联历史设计 | [Native v1 Turn Trace](native-v1/n8-turn-trace.md)、[Trace 数据源缺口](native-v1/n8-backend-data-contract.md) |
| 版本关系 | 首次设计，把 Native v1 Trace 设计移植到 React Desktop |

结论：右上角 capabilities/artifacts 入口已删除，替换为 Turn trace 入口。React overlay 默认选择当前 Turn 并自动执行 Trace 验证。普通会话真实 Trace 仍受 Runtime 数据源限制，界面显示真实 404 原因，不伪造数据。

实现状态：入口替换、React overlay、Go Trace 读取绑定、默认验证、404 边界和测试已完成。普通会话真实 Trace 数据等待后端 owner 路由与采集策略。

## 2. 背景、问题与证据

### 2.1 当前问题或机会

React Desktop 右上角入口原本打开 capabilities/artifacts。用户需要直接查看当前对话 Turn 的执行轨迹，包括 summary、模型/工具步骤、耗时、coverage 和错误。arkclaw-admin 已有成熟的 Trace 页面，可以作为交互和视觉参考。

### 2.2 已验证事实

| ID | 类型 | 结论/假设 | 精确证据 | 设计影响 | 验证/关闭方式 |
|---|---|---|---|---|---|
| FACT-1 | 已验证事实 | arkclaw-admin 的 Trace 页面包含 summary、coverage、Turn 选择、瀑布图、层级列表、筛选和 step 详情。 | `arkclaw-admin/frontend/src/components/enterprise/EnterpriseAgentTrajectory.vue:1`；`frontend/src/utils/enterprise-agent-runtime.js:896`。 | React overlay 对标该信息层级。 | 源码复核。 |
| FACT-2 | 已验证事实 | arkclaw-admin 通过 Runtime Console Trace 接口读取数据。 | `frontend/src/components/enterprise/EnterpriseAgentConsoleRoute.vue:480`。 | OpenMPA 需要 Go 绑定读取同一 Runtime 数据，WebView 不持有 Runtime key。 | 源码复核。 |
| FACT-3 | 已验证事实 | OpenMPA Native v1 已经设计并实现过 Turn Trace，包括右上角入口、Turn 筛选、summary、时间轴、缓存和 404 边界。 | `prd-spec/features/agentkit-desktop/native-v1/n8-turn-trace.md:1`。 | React 版复用同一产品语义。 | 文档复核。 |
| FACT-4 | 已验证事实 | 普通 owned Session 的 Trace 当前返回 404，Runtime 只接受 admin_debug Session，普通 Turn 默认不采集。 | `agentkit-mpa-agent/app/runtime_console/router.py:150`、`:159`；`app/services/session/crud.py:428`；`app/stores/session_meta.py:99`。 | 普通会话 Trace overlay 必须显示真实不可用状态。 | 已有 fixture 证据与源码复核。 |
| FACT-5 | 已验证事实 | 右上角原按钮在 `desktop/internal/host/rail.go` 的 titlebar，原 action 为 `artifacts`。 | `desktop/internal/host/rail.go:198`（原实现）；替换后为 `trace`。 | 按钮改为打开 React Trace overlay。 | 代码与测试。 |
| FACT-6 | 已验证事实 | React Trace overlay 默认选择当前 invocation 并自动读取。 | `src/client/desktop/TracePanel.tsx:1`；`src/client/desktop/Chat.tsx:526`。 | 满足“默认开启 trace 验证”。 | 组件测试。 |
| FACT-7 | 已验证事实 | Go Trace 绑定读取 `/api/v1/runtime-console/sessions/{sessionId}/trace?invocationId=`，并校验 session/invocation 与 step 数量。 | `desktop/internal/trace/service.go:95`。 | 凭据留在 Go，WebView 只拿规范化结果。 | 服务测试。 |
| FACT-8 | 已验证事实 | Trace 详情只显示经过敏感键过滤和长度限制的 JSON。 | `src/client/desktop/trace-model.ts` 的 `safeTraceDetails`。 | 避免 Authorization/token/secret 进入 UI。 | 单元测试。 |

### 2.3 假设与待确认项

| ID | 类型 | 结论/假设 | 精确证据 | 设计影响 | 验证/关闭方式 |
|---|---|---|---|---|---|
| OPEN-1 | 待确认项 | 普通会话 owner Trace 路由和采集策略何时可用。 | 当前无普通会话 Trace 数据。 | 决定何时能把 404 空状态替换为真实数据。 | 后端合同落地并在授权 Runtime 验收。 |
| OPEN-2 | 待确认项 | 真实 GUI 截图验收执行时间。 | 已安装 OpenMPA 占用单实例锁。 | 决定视觉验收能否关闭。 | 关闭应用后运行开发构建。 |

## 3. 目标、非目标与范围

### 3.1 目标

1. 右上角入口从 artifacts 切换为 Turn trace。
2. Trace overlay 默认当前 Turn，自动执行 Trace 验证。
3. 展示 summary、coverage、Turn 选择、搜索、状态/类型筛选、timeline/tree 和 step 详情。
4. 普通会话 404 时显示明确原因，不伪造 Trace。
5. Trace 读取走 Go 绑定，不在 WebView 暴露 Runtime key。

### 3.2 非目标

- 不在本期新增 Runtime 后端 Trace 路由或采集策略。
- 不把普通 Session 改为 admin_debug。
- 不估算缺失的 Token、成本或耗时。
- 不保留 artifacts 入口。
- 不发布新版本。

### 3.3 范围

范围内：React Desktop titlebar 入口、Trace overlay、Go Trace 读取、默认验证、404 边界、测试。

范围外：Runtime Console 后端、普通会话采集、APM 补图、发布流水线。

## 4. 用户、调用方与场景

| 场景 | 角色 | Given | When | Then |
|---|---|---|---|---|
| 查看当前 Turn | 桌面用户 | 已有一个 accepted Turn | 点击右上角 Trace | overlay 打开，默认当前 Turn，自动显示 summary 和步骤 |
| 查看历史 Turn | 桌面用户 | 同一 Session 有多个 invocation | 切换 Turn 下拉 | 只展示选中 invocation 的 Trace |
| 普通会话无 Trace | 桌面用户 | Runtime 只记录 admin_debug | 打开 Trace | 显示 admin_debug 限制，不显示假步骤 |
| 运行中 Turn | 桌面用户 | 当前 Turn running | 打开 Trace | coverage 显示 active/collecting，已观测步骤可见 |
| 搜索/筛选 | 桌面用户 | Trace 已加载 | 输入搜索或选择状态 | 只保留匹配步骤，保留匹配子步骤的父节点 |

## 5. 需求与约束

### 5.1 功能需求

| ID | 来源 | 优先级 | 触发/适用范围 | 必须行为或约束 | 可观察结果/度量 |
|---|---|---|---|---|---|
| FR-1 | 用户原始请求 / FACT-5 | 必须 | titlebar | 删除 artifacts 入口，右上角按钮打开 Trace overlay。 | 按钮标签为 Open turn trace；点击后 overlay 可见。 |
| FR-2 | 用户原始请求 / FACT-6 | 必须 | 打开 Trace | 默认选择当前 invocation，并自动读取/验证。 | 无需手动点击刷新即可看到 loading/ready/error。 |
| FR-3 | 用户原始请求 / FACT-1 | 必须 | Trace overlay | 显示 summary、coverage、Turn 选择、搜索、状态/类型筛选、timeline/tree、step 详情。 | 组件测试找到对应控件与文本。 |
| FR-4 | 用户原始请求 / FACT-4 | 必须 | 普通会话 404 | 显示 admin_debug 限制和不可用原因。 | 404 测试文本包含 admin_debug。 |
| FR-5 | 用户原始请求 / FACT-2 | 必须 | Trace 读取 | 通过 Go 绑定读取，校验 session/invocation 和 step 数量。 | 服务测试覆盖成功与 404；WebView 不持有 Runtime key。 |
| FR-6 | 安全约束 / FACT-8 | 必须 | step 详情 | 敏感键过滤、长度限制、不展示原始 Authorization/token。 | `safeTraceDetails` 测试通过。 |
| FR-7 | 用户原始请求 | 必须 | 全部改动 | 现有测试、类型检查和构建通过。 | Go、TS、Vitest、Vite build 全部通过。 |

### 5.2 非功能约束

| ID | 来源 | 优先级 | 触发/适用范围 | 必须行为或约束 | 可观察结果/度量 |
|---|---|---|---|---|---|
| NFR-1 | FACT-7 | 必须 | Trace 读取 | 30 秒超时，最多 1000 steps，拒绝不匹配 session/invocation。 | 服务测试与代码常量。 |
| NFR-2 | 用户原始请求 | 必须 | UI | 打开/关闭不重建窗口，不丢草稿；默认无入场动画，颜色过渡 150ms。 | CSS 与组件测试。 |
| NFR-3 | `desktop/README.md` 凭据合同 | 必须 | Trace 读取/详情 | 不把 Runtime key、JWT、Authorization 或原始 secret 写入 shell state、日志或 UI。 | 代码审查与安全测试。 |
| NFR-4 | 用户原始请求 | 必须 | 全部改动 | 不新增运行时依赖。 | 构建与测试。 |

## 6. 方案设计

### 6.1 方案总览

React Desktop 保留原生 titlebar，只把右侧按钮 action 从 `artifacts` 改为 `trace`。React 收到 `trace` 后切换 `traceOpen`，在 WebView 内渲染全屏 Trace overlay。Trace 数据由 Go `trace.Service` 读取 Runtime Console 接口，React 只接收规范化后的 `TraceVM`。overlay 默认当前 Turn 并自动读取，404 时显示真实原因。

### 6.2 关键流程、状态与不变量

1. 原生 titlebar 按钮点击，发出 `trace` action。
2. `DesktopAppState` 切换 `traceOpen`，并写回 ShellState。
3. `DesktopChat` 从 projection messages 提取最近 5 个 invocation，当前 invocation 优先。
4. `TracePanel` 默认选择当前 invocation，调用 `services.trace.read`。
5. 成功时构建树、summary、coverage 和 timeline。
6. 失败时清空 trace，显示错误；404 特判为 admin_debug 限制。
7. 关闭 overlay 只改 `traceOpen`，不改变 Session、草稿或窗口。

不变量：Trace 只属于精确 session + invocation；不按文本或时间窗猜测；缺失 metrics 显示 `—`，不显示 0；敏感详情不进入 UI。

### 6.3 接口、数据与配置契约

Go 绑定：

```json
{"connectionEpoch": 3, "sessionId": "session-1", "invocationId": "inv-1"}
```

Runtime 路径：

```text
GET /api/v1/runtime-console/sessions/{sessionId}/trace?invocationId={invocationId}
```

返回 `traceVersion`、`sessionId`、`invocationId`、`run`、`coverage`、`steps`。客户端校验 `sessionId` 与 `invocationId` 回显一致，steps 不超过 1000。

ShellState 新增可选 `traceOpen`。React 发送该字段，原生 titlebar 用它显示选中状态。

### 6.4 备选方案与取舍

| 方案 | 正确性 | 复杂度 | 兼容性 | 选择 |
|---|---|---|---|---|
| React overlay + Go Trace 绑定 | 复用 React UI，凭据留在 Go | 中 | 不改变原生 titlebar | 采用 |
| 原生 traceui overlay | 已有 Native 实现 | 高 | 需要移植 application/traceui | 放弃本期 |
| WebView 直接请求 Runtime | 会暴露 Runtime key | 低 | 不安全 | 放弃 |
| 只保留 artifacts | 不满足用户目标 | 低 | 无 | 放弃 |

## 7. 边界、失败与恢复

| 场景 | 系统行为 | 状态是否改变 | 可观察信号 | 恢复方式 |
|---|---|---|---|---|
| 无 accepted Turn | 显示 Waiting for an accepted turn | 否 | empty state | 发送消息后自动选择 |
| 404/sessionNotFound | 显示 admin_debug 限制 | 否 | error banner | 等待后端 owner 路由 |
| 401/403 | 显示身份无权限 | 否 | error banner | 重新认证 |
| 断连/epoch 变化 | Go 返回 staleConnection/notConnected | 否 | error banner | 重新连接后刷新 |
| Trace 为空 steps | 显示 not available | 否 | empty state | 刷新或换 Turn |
| 详情含敏感键 | 过滤后显示 | 否 | safe details | 无 |
| 关闭 overlay | 只关闭面板 | 否 | chat 保持 | 再次打开 |

## 8. 兼容、迁移与回滚

- 删除 artifacts 入口与 pane，原 `Artifacts.tsx`、native `artifacts.go` 删除。
- `ShellState` 删除 `pane`，新增可选 `traceOpen`。旧 WebView 不发送 `traceOpen` 时按 false。
- Trace 绑定是新增绑定，旧 WebView 不调用不影响。
- 回滚代码即可恢复 artifacts 入口；无持久化数据变更。

## 9. 安全、隐私与权限

- Runtime key 只在 Go 连接层；WebView 不持有。
- Trace 读取校验 session/invocation 回显，拒绝跨 Session 响应。
- 详情过滤 Authorization/token/secret/password/apiKey/cookie 键。
- 不把原始 prompt、工具参数、JWT 写入 shell state 或日志。
- 普通会话 404 不降级为 admin_debug，不绕过权限。

## 10. 可观测性、发布与运维

- UI 显示 Trace validation 状态：loading、ready、error。
- 404 文案包含 admin_debug 限制，便于用户和工程师定位数据源。
- 无新增远端指标。发布检查为 Go/React 测试、类型检查、构建。

## 11. 实施计划与影响范围

| 阶段 | 变更边界 | 完成证据 |
|---|---|---|
| 1 | Go Trace 读取绑定 | `go test ./internal/trace` |
| 2 | titlebar 入口替换，删除 artifacts pane | `go test ./internal/host ./internal/nativeui` |
| 3 | React Trace overlay 与默认验证 | `npx vitest run tests/desktop-trace-panel.test.tsx` |
| 4 | Trace 归一化/安全详情 | `npx vitest run tests/trace-model.test.ts` |
| 5 | 全量门禁 | Go、TS、Vitest、build |

影响文件：`desktop/internal/trace`、`desktop/internal/host/rail.go`、`desktop/internal/nativeui/controller.go`、`src/client/desktop/TracePanel.tsx`、`src/client/desktop/trace-model.ts`、`src/client/desktop/trace.css`、`src/client/desktop/Chat.tsx`、`src/client/desktop/DesktopApp.tsx`、`src/client/desktop/DesktopAppState.tsx`。

## 12. 验收标准与验证计划

| ID | Given | When | Then | 验证方法/命令 | 环境 |
|---|---|---|---|---|---|
| AC-1 | 有 accepted Turn | 点击右上角 Trace | overlay 打开并默认当前 Turn | `npx vitest run tests/desktop-navigation.test.tsx tests/desktop-trace-panel.test.tsx` | 本地 |
| AC-2 | Trace 成功返回 | 打开 overlay | 显示 summary、coverage、步骤和详情 | `npx vitest run tests/desktop-trace-panel.test.tsx` | 本地 |
| AC-3 | 普通会话 404 | 打开 overlay | 显示 admin_debug 限制 | `npx vitest run tests/desktop-trace-panel.test.tsx` | 本地 |
| AC-4 | Trace payload 两页/嵌套 | 构建模型 | 树、summary、时间条和过滤正确 | `npx vitest run tests/trace-model.test.ts` | 本地 |
| AC-5 | 敏感详情 | 打开 step 详情 | Authorization/token 被过滤 | `npx vitest run tests/trace-model.test.ts` | 本地 |
| AC-6 | 全量代码 | 运行门禁 | Go、TS、Vitest、build 通过 | `go test ./...`；`npx tsc --noEmit`；`npx vitest run`；`npm run build:desktop:frontend` | 本地 |

## 13. 风险与待确认项

| ID | 触发条件 | 概率/影响 | 缓解措施 | 监测信号 | 兜底/回滚 |
|---|---|---|---|---|---|
| RISK-1 | 普通会话 Trace 持续 404 | 高/高 | 明确空状态；等待后端 owner 路由 | 用户反馈 | 保留 artifacts 代码回滚能力（Git） |
| RISK-2 | Trace 详情含未脱敏字段 | 中/高 | 客户端敏感键过滤和长度限制 | 安全测试 | 关闭详情或只显示白名单 |
| RISK-3 | 大 Trace 渲染卡顿 | 中/中 | 最多 1000 steps，CSS 列表，搜索过滤 | 性能观察 | 虚拟列表后续优化 |
| RISK-4 | 真实 GUI 未验收 | 中/中 | 组件测试与 CSS 断言 | 发布前人工验收 | 关闭已安装应用后运行开发构建 |

| ID | 问题 | 为什么重要 | 影响的决策 | 确认方式 | 最晚确认阶段 |
|---|---|---|---|---|---|
| OPEN-1 | 普通会话 owner Trace 路由何时可用 | 决定真实数据能否出现 | FR-2/FR-4 | 后端合同落地 | 下一轮真实数据验收前 |
| OPEN-2 | 真实 GUI 截图验收执行人 | 决定视觉验收关闭 | AC-1/AC-2 | 关闭应用后运行开发构建 | 发布前 |

## 14. 追踪矩阵

| 需求 | 方案位置 | 验收标准 | 验证方法 | 状态 |
|---|---|---|---|---|
| FR-1 | 6.2 / titlebar | AC-1 | `desktop-navigation.test.tsx` | 已覆盖 |
| FR-2 | 6.2 / TracePanel | AC-1 | `desktop-trace-panel.test.tsx` | 已覆盖 |
| FR-3 | 6.2 / TracePanel | AC-2 | `desktop-trace-panel.test.tsx` | 已覆盖 |
| FR-4 | 6.2 / 404 | AC-3 | `desktop-trace-panel.test.tsx` | 已覆盖 |
| FR-5 | 6.3 / Go binding | AC-1、AC-6 | `go test ./internal/trace` | 已覆盖 |
| FR-6 | 9 / safe details | AC-5 | `trace-model.test.ts` | 已覆盖 |
| FR-7 | 11 / gates | AC-6 | 全量门禁 | 已覆盖 |
| NFR-1 | 6.3 | AC-6 | 服务测试 | 已覆盖 |
| NFR-2 | 6.2 | AC-1 | 组件测试 | 已覆盖 |
| NFR-3 | 9 | AC-5 | 安全测试 | 已覆盖 |
| NFR-4 | 11 | AC-6 | 构建 | 已覆盖 |
