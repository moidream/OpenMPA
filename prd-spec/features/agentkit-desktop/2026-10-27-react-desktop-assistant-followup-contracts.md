# React Desktop 助手、定时任务与主题：接口与配置契约

### 6.3 接口、数据与配置契约

`Host.Theme` 返回：

```json
{ "ok": true, "value": { "theme": "dark" } }
```

`theme` 只允许 `system`、`light`、`dark`。未知值按 `system` 处理。

`Agents.Update` 请求：

```json
{
  "connectionEpoch": 3,
  "mpaInstanceId": "agent-1",
  "agentsMd": "...",
  "skills": [
    { "id": "skill-1", "name": "Skill one", "type": "skill", "enabled": true }
  ]
}
```

`skills` 最多 200 条，`id` 与 `name` 最多 256 字符，不允许换行和 NUL。

`Schedule.ListForAgent` 返回合成 `TaskPage`：

```json
{ "items": [], "limit": 200, "offset": 0, "countAvailable": false }
```

`items` 只包含当前 `agentId` 的任务。`total`、`overview`、`hasMore`、`nextOffset` 不设置。

配置契约：`display.theme` 新增键，默认 `system`；现有数据库不需要 schema 升级。

### 6.4 备选方案与取舍

| 方案                        | 正确性                     | 复杂度 | 兼容性       | 运维成本         | 选择                   |
| --------------------------- | -------------------------- | ------ | ------------ | ---------------- | ---------------------- |
| 客户端跨页过滤              | 能覆盖当前 API，跨页不漏   | 中     | 不依赖后端   | 最多读取 1000 条 | 采用                   |
| 只过滤当前页                | 会漏任务                   | 低     | 无           | 低               | 放弃                   |
| 等待后端新增 `agentId` 参数 | 最省客户端逻辑             | 低     | 需要后端发布 | 中               | 作为 OPEN-1 的后续选项 |
| skill space 本地 JSON 导入  | 可演示但不等同 skill space | 低     | 会产生假能力 | 低               | 放弃                   |
| 保留 Spaces 死代码          | 不影响运行但继续维护       | 低     | 无           | 中               | 放弃                   |
