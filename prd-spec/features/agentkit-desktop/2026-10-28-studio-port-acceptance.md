# Studio 功能迁移本地验收

日期：2026-10-28。结论：本地实现和验收通过。

## 来源与适配

用户所述 `codex/main-20261026` 在本地和远端均未找到。
按日期和功能确认来源为 `codex/work-20261026`，提交 `7f9fe23`。
目标为当前 `codex/main-20261028`，HEAD 保持 `f7c7f95`。

迁移沿用当前 React 桌面架构、现有连接和会话接口。
原生 rail 新增资源与协作入口，React 显示内容。
保留任务开始前的未提交汉化修改及用户目录 `.trae-html-share-packages/`。
未提交、推送、部署或写入云端。

## 功能结果

| 能力 | 当前行为 | 验收证据 |
| --- | --- | --- |
| Skill | 只显示当前 Runtime mounts 中的技能空间；签名读取目录、本地分页；选择加入会话草稿 | Go 资源测试、React 资源测试、浏览器 Skill 选择 |
| MCP | 显示所选会话的服务及工具说明；区分未提供工具列表与空列表；连接现有会话设置 | MCP 过期响应测试及浏览器工具展示 |
| 共享 AGENTS.md | 读取正文及 revision，显式提交版本化修改；冲突后保留编辑并要求读回 | Unicode 边界、冲突、丢失响应、浏览器冲突与语言切换 |
| Main Profile | 使用会话实例身份或现有 Runtime overview 身份；Go 保存一次性快照；仅更新指令 | 快照篡改、并发重放、scope 切换、404 不初始化与浏览器编辑 |
| A2A | 分页读取空间和成员；确认后绑定、解绑与更新别名；仅运行中且别名可用的成员可选择 | CAS 请求、绑定前后校验、成员不可用回归及浏览器绑定/别名 |
| 会话草稿 | Skill 追加，@ 任务前置，均保留原有文本；不自动发送；回执只更新提交对应的草稿 | 跨页面回执、后续选择、手工编辑、会话/身份隔离与浏览器 |
| 中英文 | 新页面与安全错误使用现有词典；切换语言保留编辑内容 | 原有汉化回归、新服务错误测试及两个编辑器的浏览器切换 |

## 检查结果

| 检查 | 结果 |
| --- | --- |
| `npm test` | 96 个文件、531 项测试通过 |
| `npm run typecheck` | 通过 |
| `npm run lint` | 通过 |
| `npm run build:desktop:frontend` | 通过 |
| `npm run desktop:generate` | 通过，词典与 IPC 生成文件一致 |
| `npm run design:check` | 通过 |
| `GOTOOLCHAIN=go1.27.3 go test -race ./...`，在 `desktop/` | 全量通过；随后实施 Owner 对修改后的测试完成定向 race 检查 |
| `GOTOOLCHAIN=go1.27.3 go vet ./...`，在 `desktop/` | 最终通过 |
| `git diff --check` | 通过 |
| 新页面、回归用例及浏览器脚本的 Prettier 检查 | 通过 |
| `node scripts/verify-desktop-studio.mjs http://127.0.0.1:4178` | 10 项检查通过，零页面错误、零缺失 IPC 方法 |

测试期间修正了新测试的返回类型、未使用 import 和取消函数清理问题；
以上记录为修正后的结果。构建仍提示现有 Vite 配置加载方式和大文件块，
均未阻断构建，本次未扩展到打包策略调整。

前端完整日志：
`/tmp/openmpa-studio-port.zAeUcK/final-ts-tests.log`。

## 浏览器证据

环境为实际 React 页面、Chrome 和模拟 MyGo IPC。
操作通过按钮、输入和下拉框完成，读取页面状态用于断言。
模拟环境不发出真实 Runtime 或云端请求。

最终结果目录：
`.openmpa/studio-verification/2026-10-28T08-10-40.669Z/`。
包含 `report.json`、`evidence.json`、会话草稿及两个新页面截图。

通过的操作包括 Skill 选择、MCP 工具展示、两个指令编辑器的冲突/读回/语言切换、
A2A 解绑与绑定、保存别名后加入 @ 草稿、不自动提交、保留页面实际隐藏，
以及两个页面在 1280、768、320 像素下的宽度和垂直滚动检查。

## 两轮审查

第一轮检查 Runtime 契约、IPC 绑定、词典及现有会话集成。
第二轮检查损坏、过期、冲突、scope 切换、快照篡改、重排回执和页面显示反例。
审查以任务开始时保存的脏工作区为基线，避免把此前的汉化改动算作迁移缺陷。

| 确认问题 | 修复与验证 |
| --- | --- |
| Profile 读取过程中 scope 改变可能将旧快照关联到新 scope | 从请求前捕获 scope，在身份解析、读取、保存及保留快照时检查；Go 回归拒绝新 scope 下的快照 |
| `.desktop-form` 的 grid 显示规则覆盖保留页面的 hidden 状态 | 增加 scoped hidden 规则并使用现有内容布局；浏览器逐次验证实际可见性和滚动 |
| 新 Usage 错误映射遗漏显式业务身份错误 | 保留机器错误字段并接入已有中英文恢复说明；服务测试覆盖 |
| 已停止成员或别名不可用的成员仍显示可用的选择按钮 | UI 与处理函数检查运行状态和 aliasAvailable；两个反例在修复前失败，修复后通过 |

StrictMode 首次点击必然阻塞的推测经原始浏览器流程排除。
浏览器修改 Profile 投影能够覆盖模型/技能的推测经私有快照和一次性 token 测试排除。
当前无未解决的确认缺陷。

审查账本为：
`/tmp/openmpa-studio-port.zAeUcK/review-ledger.json`。
已使用 `code-review/scripts/validate_review_ledger.py` 验证，结果为 `pass`。

## 本地验收边界

真实 Runtime 的权限、版本差异，以及共享指令、Profile、A2A 的云端写入，
仍需连接对应环境后验证。没有能力或没有配置的旧 Runtime 显示能力不可用，
继续使用原有聊天功能。404/405 不触发 Profile 初始化。

资源编辑和会话草稿保存在本次 App 生命周期内。
本文不承诺进程退出或崩溃后的草稿持久化恢复。
