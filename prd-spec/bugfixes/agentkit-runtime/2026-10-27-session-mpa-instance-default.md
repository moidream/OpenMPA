# [Runtime][P1] 新建普通 Session 未持久化默认 MPA Agent ID，导致 execution-config 404 和模型无法切换

## 1. 基本信息

| 字段 | 内容 |
|---|---|
| Bug 标题 | 新建普通 Session 未持久化默认 MPA Agent ID，导致 execution-config 404 |
| 严重级别 | P1，模型切换不可用 |
| 类型 | Runtime Session 创建缺陷 |
| 发现日期 | 2026-10-27 |
| Runtime | `r-yexsqsrpxcxfapor9fvl` |
| Runtime revision | 42 |
| 客户端 | OpenMPA React Desktop 0.1.1 |

## 2. 问题摘要

OpenMPA 通过 `POST /api/v1/sessions` 创建普通 Session 时没有传 `mpaInstanceId`。Runtime 将 `session_meta.mpa_instance_id` 持久化为空字符串。随后：

- `GET /api/v1/sessions/{id}/execution-config` 返回 404。
- Composer 无法读取 execution-config，也无法写模型 override。
- 模型下拉可以看到 Runtime catalog，但不能应用选择。

Runtime 环境已有默认 agent ID：

```text
MPA_AGENT_ID=mi-2de0361fbd2e5b1889aac5d5
```

但新建 Session 没有把它写入 `session_meta.mpa_instance_id`。

## 3. 证据

数据库 `session_meta`：

```text
session_id=88695ff5-1409-4062-b257-5ceadaa4fd21
session_kind=normal
mpa_instance_id=''
profile_revision=0
execution_config_revision=0
trace_enabled=false
```

同 Runtime 中已有正确 Session：

```text
session_id=b3c54794-bc1a-42a8-bafc-c61c8f2cd83a
mpa_instance_id=mi-2de0361fbd2e5b1889aac5d5
execution_config_revision=2
```

```text
session_id=3ce65800-7862-4dbd-ba8e-9ad0a6a83041
mpa_instance_id=mi-2de0361fbd2e5b1889aac5d5
execution_config_revision=2
```

Runtime 日志：

```text
GET /api/v1/sessions/88695ff5-1409-4062-b257-5ceadaa4fd21/execution-config 404
GET /api/v1/agents 404
```

Runtime 环境：

```text
MPA_AGENT_ID=mi-2de0361fbd2e5b1889aac5d5
```

源码：

- `app/services/session/crud.py` 的 `create_session` 和 `create_or_get_session` 接受 `mpa_instance_id`，但没有在为空时回退到 `settings.mpa_agent_id`。
- `app/services/session/execution_config.py` 的 `get_session_execution_config` 在 `meta.mpa_instance_id` 为空时抛出 `SessionExecutionConfigNotFound`。

## 4. 影响

- 新建普通 Session 无法读取或修改 execution-config。
- 模型切换、MCP override 等 execution-config 能力全部不可用。
- 用户只能看到 Runtime catalog，但无法选择模型。
- 旧 Session 也无法通过客户端修复，因为 `mpa_instance_id` 不可更新。

## 5. 期望行为

`POST /api/v1/sessions` 未传 `mpaInstanceId` 时，Runtime 应使用当前默认 `MPA_AGENT_ID`：

```text
mpa_instance_id = settings.mpa_agent_id
```

并持久化到 `session_meta`，使 `GET /execution-config` 能创建初始 revision。

如果默认 agent ID 不存在，应返回明确错误，而不是创建空 `mpa_instance_id` 的 Session。

## 6. 建议修复

1. 在 `create_session` / `create_or_get_session` 中，当 `mpa_instance_id` 为空时回退到 `settings.mpa_agent_id`。
2. 在 `session_meta` 写入该默认值。
3. 确保首次 `GET /execution-config` 能通过 `create_initial` 创建 revision。
4. 对历史空 `mpa_instance_id` Session 提供受控回填方案，或明确标记不可用。
5. 增加回归测试：新建 Session 后 execution-config 200，模型 override 可读写。

## 7. 验收

| ID | Given | When | Then |
|---|---|---|---|
| AC-1 | Runtime 配置了 `MPA_AGENT_ID` | `POST /sessions` 不带 `mpaInstanceId` | `session_meta.mpa_instance_id` 等于默认 agent ID |
| AC-2 | 新建普通 Session | `GET /execution-config` | 返回 200 和初始 revision |
| AC-3 | 新建普通 Session | `PATCH /execution-config` 设置 model | 返回新 revision，后续 Turn 使用该模型 |
| AC-4 | 历史空 agent ID Session | 读取 execution-config | 明确 404 或可恢复状态，不返回空成功 |

## 8. 客户端临时规避

- OpenMPA 普通“新建对话”不再创建 `admin_debug` Session。客户端从 Runtime Console overview 读取当前 agent ID，并创建绑定该 ID 的普通 Session，使每个被接受的 Turn 固化自己的 execution-config revision。
- 对没有 execution-config 的旧普通 Session，打开会话不会产生远端写入。用户第一次明确选择模型时，客户端用精确 Session ID 和当前 agent ID 补齐绑定，再执行 CAS patch。
- 当前 Turn 处于 running、submitting、结果未知或停止核验阶段时仅锁住下一 Turn 的选择；权威终态到达后立即恢复，并可切回 Runtime default。
- 历史 `admin_debug` Session 不能通过普通 Session API 无损改变类型。选择模型时显示该精确限制，不再把它概括成“此对话不能切换模型”或要求用户通过新建对话才能选择。

## 9. 终止竞态补充

模型锁定和 Stop 必须共享权威 Turn 状态。客户端发出 abort 后关闭旧订阅，并用 exact invocation 对 history、activeRun 与 pending receipt 做最多四轮有界核验：

- `aborted=false` 但 history 已 terminal，按完成收敛；
- `aborted=true` 且 terminal 稍后到达，保持核验态直到收敛；
- Runtime 仍明确报告 active，恢复 running 并提示 abort 未确认；
- 超过有界核验仍无终态，进入 unknown，不误开放下一 Turn。
