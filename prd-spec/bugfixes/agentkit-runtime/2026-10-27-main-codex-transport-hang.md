# [Runtime][P1] Main Codex transport 异常后 SSE 挂起 10 分钟且丢失底层错误

## 1. 基本信息

| 字段 | 内容 |
|---|---|
| Bug 标题 | Main Codex transport 异常后 SSE 挂起 10 分钟且丢失底层错误 |
| 严重级别 | P1，交互式对话长时间无响应 |
| 类型 | Runtime 缺陷与诊断缺口 |
| 发现日期 | 2026-10-27 |
| 运行时 | `r-yexsqsrpxcxfapor9fvl` |
| Runtime revision | 42 |
| Codex 版本 | `0.159.2` |
| Session | `88695ff5-1409-4062-b257-5ceadaa4fd21` |
| 失败 Invocation | `e-b7a643a8-82ba-4b5e-bcde-f068b7ca482a` |
| Thread | `01a17bfb-7362-7b30-94a7-04165836938e` |
| 客户端 | OpenMPA React Desktop 0.1.1 |

## 2. 问题摘要

用户发送一条普通对话后，Runtime 在 06:10:54 发生 Main Codex transport 读取异常。Runtime 将 Turn 标记为 `reconciliation_required`，保持 Run 非终态，并且不向 SSE 客户端发布 done 或 error 事件。SSE 连接持续 10 分 14 秒，客户端一直显示“正在提交/接收”和 `running`，没有流式输出，也没有终态。

用户最终触发停止后，Runtime 在 06:21:00 以 `HOST_CANCEL_AUTHORITY_ACCEPTED` 结束该 Turn，SSE 在 06:21:01 关闭。

## 3. 影响

- 普通交互式 Turn 在 Main Codex transport 异常后可能挂起最多 600 秒。
- 客户端无法区分“模型仍在生成”和“Runtime 已进入 reconciliation”，因为 SSE 没有发送中间状态或终态。
- 用户只能手动停止，且停止前没有任何可解释的错误。
- 底层 transport 异常类型、Codex stderr、进程退出码都没有记录，排障只能靠时间线推断。
- 同类异常可能影响所有使用 Main Codex 的普通会话。

## 4. 复现

### 4.1 前置条件

- 使用 Runtime `r-yexsqsrpxcxfapor9fvl`，revision 42。
- 使用普通 Session，不是 admin_debug Session。
- 客户端发送一条普通对话。

### 4.2 步骤

1. 在 Session `88695ff5-1409-4062-b257-5ceadaa4fd21` 发送：
   `介绍下你的工作环境配置`
2. 观察 Runtime 日志和 SSE 连接。
3. 等待 10 分钟或手动停止。

### 4.3 实际结果

- 06:10:47 Runtime 接收 `/run`，返回 invocation `e-b7a643a8-82ba-4b5e-bcde-f068b7ca482a`。
- 06:10:47 SSE 连接建立，HTTP 200。
- 06:10:49 Codex `turn/start` 成功。
- 06:10:49 model relay 返回 `firstTokenMs=2182.597`。
- 06:10:54 Codex transport `stdout` 报 `invalid_or_closed_stream`。
- 06:10:54 model relay 报 `ClientConnectionResetError` / `downstream_disconnect`。
- 06:10:54 Runtime 报 `reconciliation_required`。
- 06:10:57 Runner 记录 `Keeping Main run nonterminal for reconciliation`。
- 06:10:57 Inbox runner 返回 `status=running should_publish_done=False`。
- 06:21:00 Runtime 记录 `HOST_CANCEL_AUTHORITY_ACCEPTED`。
- 06:21:01 SSE 请求在 614 秒后关闭。
- 客户端全程没有收到 done/error，也没有收到可见的流式文本。

### 4.4 期望结果

Transport 异常后，Runtime 应在合理时间内向 SSE 客户端发布可判定的终态或明确的中间状态，例如：

- 立即发布 `reconciliation_required` 事件，让客户端显示“结果待确认”。
- 或者关闭 SSE 并发布 error，客户端可以提示用户重试或读取历史。
- 不应该让客户端在无事件的情况下等待 600 秒。

## 5. 证据

### 5.1 Runtime 主日志

```text
06:10:47.480 Enqueued run invocation_id=e-b7a643a8-82ba-4b5e-bcde-f068b7ca482a
06:10:47.569 POST /api/v1/sessions/88695ff5-1409-4062-b257-5ceadaa4fd21/sse 200
06:10:49.589 codex_transport pid=19946 rpcMethod=config/read status=success
06:10:49.599 codex_transport pid=19946 rpcMethod=thread/resume status=success
06:10:49.679 codex_transport pid=19946 rpcMethod=turn/start status=success
06:10:49.909 runner step=first_event partial=false
06:10:54.749 codex_transport pid=19946 step=stdout status=failed errorType=invalid_or_closed_stream
06:10:54.799 model_relay errorType=ClientConnectionResetError failureStage=downstream_disconnect
06:10:54.819 controller status=reconciliation_required step=consume_turn
06:10:54.909 controller status=reconciliation_required step=turn_complete durationMs=5842.275
06:10:57.039 runner Keeping Main run nonterminal for reconciliation
06:10:57.209 inbox_runner status=running should_publish_done=False
06:21:00.019 controller completion_arbiter_decision reason=HOST_CANCEL_AUTHORITY_ACCEPTED
06:21:01.919 POST /api/v1/sessions/.../sse request canceled elapsed=10m14.363s
```

### 5.2 Codex thread history

从 Codex SQLite `thread_history_1.sqlite` 读取：

```text
thread_id=01a17bfb-7362-7b30-94a7-04165836938e
turn_id=01a17bfb-9ccf-7270-b502-5657dc7405dc
status=interrupted
error_json=None
started_at=1793081449
completed_at=1793081454
duration_ms=5084
first_user_item_id=01a17bfb-9cd7-7ee0-a827-5cc8773c44fe
final_agent_item_id=None
```

这说明 Codex 收到了该 Turn，但在 06:10:54 被中断，没有持久化 agent 回复。`status=interrupted` 与 Runtime 在 transport 异常后调用 `transport.close()` 的行为一致。没有证据表明 Codex 进程先崩溃。

### 5.3 Codex 进程

- 失败前 Codex 进程为 pid `19946`。
- 06:20:59 新 Codex 进程 pid `60929` 启动，并成功 `thread/resume` 同一 thread。
- 失败进程 `19946` 在检查时已不存在。
- 这更符合 Runtime 关闭旧 transport 后重建 resident host，而不是 Codex 自行崩溃。

### 5.4 资源与崩溃检查

- cgroup `memory.events` 显示 `oom_kill 0`。
- dmesg 没有 Codex 进程崩溃或 segfault 记录。
- Codex logs DB 中 pid `19946` 在 06:10:50 到 06:10:56 之间没有任何 WARN/ERROR/INFO 记录。
- 没有 core dump。

### 5.5 模型

Codex 日志在 06:10:39 警告：

```text
Unknown model gpt-5.6-luna is used. This will use fallback model metadata.
```

这是独立配置问题，不能直接解释 transport 读取失败，但说明当前模型不在 catalog 中。

## 6. 根因分析

### 6.1 已确认的 Runtime 行为缺陷

1. **Run 非终态且 SSE 不发布终态**
   - `runner.py` 在 `MainRuntimeReconciliationRequired` 时设置 `reconciliation_required=True`。
   - `should_publish_done` 条件明确包含 `and not reconciliation_required`。
   - 因此 Runner 不会发布 done，也不会发布 error。
   - `inbox_runner.py` 记录 `should_publish_done=False`，SSE 保持打开。
   - `main_codex_reconciliation_deadline_seconds` 默认 600 秒，导致客户端最长等待 10 分钟。

2. **Transport 读取异常丢失底层原因**
   - `codex_transport.py:664` 的 `except Exception` 只记录 `errorType=invalid_or_closed_stream`。
   - 没有记录实际异常类型、异常消息、出错行的长度/前缀、Codex stderr、进程 returncode。
   - 无法区分 JSON 解析失败、通知 params 非法、notification capacity、pipe 读取失败或进程退出。
   - 这使得上游无法定位真实根因。

3. **Transport 异常后主动关闭 Codex，导致 Turn 被标记 interrupted**
   - `controller.py` 在 `except CodexTransportError` 且无可用 recovery transport 时调用 `await transport.close()`。
   - 随后抛出 `MainRuntimeReconciliationRequired("Main Codex turn stream ended before terminal event")`。
   - Codex thread 状态为 `interrupted`，没有 agent 回复，符合该路径。

### 6.2 尚未确认的根因候选

Transport `stdout` 读取异常的真实触发点尚未确认。候选包括：

- Codex app-server 输出了非 JSON 行或 malformed JSON。
- 某条 notification 的 `params` 不是 dict，触发 `_read_stdout` 的 `ValueError`。
- notification 数量或消息大小超过 transport 限制。
- 动态 tool call 请求参数非法，触发 `_handle_dynamic_tool_call_request` 的 `ValueError`。
- pipe 读取异常。

由于 `invalid_or_closed_stream` 没有保留底层异常，以上候选无法进一步收敛。需要 Runtime 先补齐诊断，才能判断是 Codex 协议问题还是 Runtime 解析问题。

## 7. 建议修复

### 7.1 P1：交互式 Turn 的 reconciliation 必须对客户端可见

- 在进入 `reconciliation_required` 时，向 SSE 发布一个明确的中间事件，例如：
  ```json
  {"kind":"reconciliation_required","terminal":false,"reason":"main_codex_transport_failed"}
  ```
- 或者关闭 SSE 并发布 error，客户端显示“结果待确认，请读取历史或停止”。
- 不要让客户端在无事件的情况下等待 600 秒。
- 对交互式 Run 使用更短的客户端可见超时，例如 30 到 60 秒；600 秒只用于后台 durable reconciliation。

### 7.2 P1：补齐 transport 诊断

在 `_read_stdout` 的异常分支记录：

- 实际异常类型和消息，去除敏感内容。
- 出错行的字节数和安全前缀，不记录完整 prompt 或 token。
- Codex 进程 pid、returncode、stderr 的安全截断内容。
- 当前 notification 数量、消息大小、最近一次 RPC 方法。
- 如果是 JSON 解析失败，记录 JSON 错误位置和行长度，不记录完整行。

### 7.3 P2：不要把可恢复的协议异常直接升级为 transport 失败

- 对单条 malformed notification，优先跳过并记录，不要立即关闭整个 Codex transport。
- 对 `params: null`、未知 notification method 等可容忍情况，返回 JSON-RPC error 或忽略，而不是让 `_read_stdout` 抛异常。
- 只有在连续协议错误、消息大小超限或进程退出时才关闭 transport。

### 7.4 P2：模型 catalog 对齐

- 将 `gpt-5.6-luna` 加入 `codex-model-catalog.json`，或把 Runtime 默认模型改为 catalog 中已知模型。
- 验证 Codex 不再输出 `Unknown model` fallback 警告。

## 8. 验收标准

| ID | Given | When | Then |
|---|---|---|---|
| AC-1 | 普通 Session 发生 Main Codex transport 异常 | SSE 客户端在线 | 60 秒内收到 `reconciliation_required` 或 error，客户端不再无限显示 running |
| AC-2 | 同一异常 | 查看 Runtime 日志 | 日志包含底层异常类型、安全消息、Codex pid/returncode/stderr 摘要 |
| AC-3 | 单条 malformed notification | Codex transport 仍在运行 | Runtime 记录并跳过该帧，不立即关闭整个 transport |
| AC-4 | 普通 Turn 正常完成 | 查看 SSE | 继续发布正常 done，不引入回归 |
| AC-5 | 模型 `gpt-5.6-luna` | Codex 启动 | 不再出现 Unknown model fallback 警告，或明确记录为已知配置 |

## 9. 临时规避

- 客户端增加流式停滞提示和“读取历史/停止”入口。
- 用户在 30 到 60 秒无事件时手动停止，避免等待 600 秒。
- 运维侧缩短 `MPA_MAIN_CODEX_RECONCILIATION_DEADLINE_SECONDS`，但需评估重复执行风险。

## 10. 待确认

1. `_read_stdout` 的真实异常类型和出错帧内容。
2. Codex app-server 是否在模型流式阶段输出了非 JSON 或 malformed notification。
3. 当前 Runtime 是否配置了 `MPA_MAIN_CODEX_RECONCILIATION_DEADLINE_SECONDS=600`。
4. 交互式 Turn 是否应该使用独立的短 reconciliation deadline。
5. `gpt-5.6-luna` 是否应加入 catalog，还是应切换到已知模型。
