# Composer 模型选择误用非 Runtime Catalog 模型

## 1. 问题

Composer 的模型选择原先是一个自由文本 Model ID 输入框。它允许用户输入任意字符串，并把该字符串作为 Session execution-config 的 `replace` 模型发给 Runtime。

实际 Runtime 日志出现了：

```text
Unknown model gpt-5.6-luna is used. This will use fallback model metadata.
```

`gpt-5.6-luna` 是 traecli 侧模型 ID，不在 Runtime 的 `codex-model-catalog.json` 中。它不是 AgentKit Runtime 支持的模型。

## 2. 证据

Runtime 的权威 catalog：

```text
/app/app/runtime/main/codex-model-catalog.json
```

catalog 中的模型包括：

- `deepseek-v4-flash`、`deepseek-v4-pro` 及其版本
- `doubao-seed-2-0-code`、`doubao-seed-2-0-lite`、`doubao-seed-2-0-pro`
- `doubao-seed-2-1-pro` 及其版本
- `glm-4-7-251222`、`glm-5-2`
- `kimi-k2`、`kimi-k2-thinking-251104`
- `dola-seed-2-1-turbo-260628`

`gpt-5.6-luna` 不在该列表中。

原 Composer 实现：

- `ComposerModelPicker.tsx` 提供自由文本输入。
- `use-composer-model.ts` 把输入原样发送到 `Session.UpdateExecutionConfig`。
- 没有 catalog 校验，也没有可筛选下拉。

## 3. 影响

- 用户可以把 traecli 或其他非 Runtime 模型 ID 写入 Session 配置。
- Runtime 会使用 fallback model metadata，行为和计费可能不可预期。
- 输入框没有模型下拉和筛选，用户无法知道哪些模型是 Runtime 支持的。
- 错误模型可能触发 Codex `Unknown model` 警告，增加运行时不稳定性。

## 4. 修复

1. 新增 `src/client/desktop/model-catalog.ts`，保存 Runtime catalog 快照。
2. 新增 `isKnownModel` 和 `filterModels`。
3. Composer 模型选择改为可筛选下拉：
   - Runtime default
   - catalog 模型列表
   - No model override
4. 当前模型不在 catalog 时显示警告：
   `Current model ... is not in the Runtime catalog. Choose a valid model before the next turn.`
5. `useComposerModel` 在 `replace` 前校验模型 ID，不在 catalog 中直接拒绝，不发网络请求。
6. 删除自由文本 Model ID 输入。

## 5. 验收

| ID | Given | When | Then |
|---|---|---|---|
| AC-1 | Runtime catalog 快照 | 打开模型菜单 | 显示可筛选模型列表，不再显示自由文本输入 |
| AC-2 | 输入 `doubao` | 筛选模型 | 只显示 Doubao 模型 |
| AC-3 | 当前模型为 `gpt-5.6-luna` | 打开模型菜单 | 显示“不在 Runtime catalog”的警告 |
| AC-4 | 选择 `doubao-seed-2-1-pro` | 应用 | 只发送 catalog 中的模型 ID |
| AC-5 | 尝试发送非 catalog 模型 | 调用 apply | 本地拒绝，不产生网络请求 |

## 6. 验证

- `npx vitest run tests/model-catalog.test.ts tests/desktop-composer-context.test.tsx`
- `npx tsc --noEmit`
- `npx vitest run --maxWorkers=1`
- `npm run build:desktop:frontend`

## 7. 剩余风险

- Runtime 目前没有公开 model catalog endpoint。客户端使用 catalog 快照，Runtime 更新 catalog 后需要同步快照。
- 当前 Runtime 的 agent config 仍可能是 `gpt-5.6-luna`。Composer 的 Session override 可以选择有效模型，但 agent config 本身仍需要 Runtime 侧或助手配置侧修正。
- 建议 Runtime 增加只读 catalog endpoint，客户端改为动态读取，删除快照。
