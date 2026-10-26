# Runtime Key Hotfix 与本地化评审

## 结论

可行性高，当前方案通过评审并已按最小边界实施。Session 修复仅删除错误的客户端身份前置条件，服务端仍是身份与权限权威；本地化复用 MyGo 单一 locale 与生成 catalog，没有建立 Go/React 双偏好状态。未修改 Runtime/Worker、远端 schema 或镜像。

## 问题与整改

| 优先级 | 维度 | 评审问题 | 影响 | 整改结果 |
|---|---|---|---|---|
| P0 | 身份契约 | v0.1.0 将可选业务身份当成 Session 必需身份 | 合法 Runtime-key-only 连接被本地拦截 | 准入改为 Connected + HasCredential；专用 Runtime 实测 CRUD 与删除回读通过 |
| P1 | IPC 最小化 | Localization.Load 若作为公开方法会被自动生成到 WebView IPC | 页面得到不需要的宿主初始化入口 | 改成 package assembly helper；生成绑定只保留 Status/Set |
| P1 | 原生状态 | 已翻译动态按钮文字曾同时作为 MyGo element key | 切换语言会丢焦点/交互状态 | 增加稳定 key 与本地化 display label 分离的 helper |
| P1 | 生命周期 | headless/worker 路径同步调用 App.Locale 会等待未运行的主循环 | 原生测试确定性超时 | 使用由 MyGo locale 事件更新的原子 locale 镜像；失败测试转绿 |
| P2 | 范围 | 高级渠道/知识/任务详细表单仍有未迁入 catalog 的文案 | 部分页签可能出现混合语言 | 当前切片保证壳层、设置、创建助手、聊天和 Space 主路径；高级表单列为后续 catalog 扩面，不冒充全量完成 |

## 验证与 GateGap

- `GG-1` 已关闭：`TestAssistantCreationAllowsRuntimeKeyIdentity` 在旧逻辑下无法编译/通过，实施后覆盖 Runtime Key、显式身份、未连接、无凭据四态，并由默认 `go test ./...` 执行。
- 真实目标 Runtime 在 `HasBusinessIdentity=false` 时完成 Session create/rename/list/delete/404 cleanup，私有证据为 `.openmpa/evidence/runtime-key-session-hotfix.json`。
- catalog 有 158 个中英文同构消息，`mygo i18n check`、生成绑定、Go 全包/vet/race、Node 76 files/384 tests、typecheck/lint/format 已通过。
- 两轮代码审查确认并修复三个问题；结构化台账 `.openmpa/evidence/runtime-key-localization-review.json` 通过 code-review validator，gate=`pass`。
- 本机外部 WindowServer 检查一度报告无窗口；同机未修改的 v0.1.0 也得到相同结果，而 LLDB 对 NSApplication 的直接只读回读显示存在 `visible=true` 的 MyGoWindow。该候选已作为环境可见性假阳性撤回，不归因于本地化代码。

## 残余范围

语言设置支持跟随系统、中文和 English，并覆盖用户完成 Runtime 选择与新建会话所需的关键路径。高级渠道、任务、知识和文档编辑器的所有细节文案尚未全部 catalog 化，因此这次不声明全应用翻译完成。该缺口不影响 Runtime-key hotfix 的正确性，也不改变既有全项目 `releaseReady=false` 状态。
