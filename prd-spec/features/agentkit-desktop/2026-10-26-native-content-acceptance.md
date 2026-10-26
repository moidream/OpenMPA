# 原生内容页验收与候选交付

公开版本已将验收账户、资源与回执标识替换为示例或占位符；实际对象仅保存在本机私有证据中。示例不是可用部署。
状态：candidate，releaseReady=false。用户要求的全模块准出尚未满足。范围是OpenMPA客户端，保持现Runtime/Worker源码与构建流程只读；使用测试账户1234567890、Runtime r-example-mpa-runtime的现OpenAPI。最新只读版本32/Ready，OpenAPI114routes。

## 已落地的界面

44px原生顶部、48px快捷栏、240px业务侧栏继续保留；头像复用原产品素材。设置、渠道、日程、知识、产物面板及assistant/Space新建使用MyGo原生ui；会话配置、改名、删除确认也使用原生对话框。聊天、富文本和本地文档编辑保留系统WebView。

账户登录/保存恢复、当前账户MPA目录分页过滤、region/project搜索切换、业务/渠道凭据、Runtime资源和独立Runtime/正式Worker Tool image表单复用现Go服务。未知操作原receipt回读；任务/知识/assistant意图在请求前按scope保存原token/内容/ID，重启不自动提交。Session目录带connectionEpoch，原生业务操作拒绝旧scope目录成员。

## 证据分层

| 验收 | 结果 | 证据 |
| --- | --- | --- |
| Node全量 | pass，75files/382tests | 本轮npm test，最近08:08执行 |
| 标准MyGo依赖Go全包、vet | pass | GOWORK=off、Go1.27.3、CGO_ENABLED=0 |
| 并发检查 | pass | nativeui/host/ipc/session/knowledge/artifact race |
| TypeScript、lint、frontend、绑定 | pass | 本轮typecheck/lint/build，MyGo generate一致 |
| 原生账户/Runtime设置与两image读取 | pass | native run20261026T000023.454249Z等；只读32与正式ToolId |
| 新Session两轮SSE、恢复、terminal/history、临时Session删除 | pass | go-live-20261025T235302Z.json |
| 原生WebView双轮对话和pending清理 | pass | native-bundle-20261026T000023.454249Z.json |
| 独立日程CAS冲突/手动成功/历史/启停/删除 | pass | go-live-20261025T235430Z.json |
| 原生日程创建/Run/启停/编辑 | pass | native-bundle-20261025T235817.380127Z.json |
| 退出客户端后Scheduled运行成功 | pass | native-scheduled-after-exit.json；精确task <private-fixture-id>，Scheduled run2b77018f-ad0f-4741-8c58-75a23ec03c58，19084ms |
| 原生assistant/改名/配置对话框/删除/Space/产物面板 | pass | native-bundle-20261026T000855.502099Z.json |
| 中文富文本/Markdown编辑、保存、持久重开 | pass | native-bundle-20261026T000208.659334Z.json |
| 原生保存/取消/知识选文件 | blocked | CGSSessionScreenIsLocked=true，blockedDesktopLocked |
| 现知识目录/导入/版本/检索/删除 | blocked | 最新目录HTTP500；Runtime32日志缺MPA_SESSION_SCOPE_SECRET，涉及OpenViking |
| 三渠道新增/更新/权限等完整操作 | blocked | 缺专用测试Bot材料；现capabilities与已有账号原生读取pass |
| 实际Runtime与Worker image更新 | not_run | 表单/原operation服务测试通过；自动验收未覆盖外部已更新的image/env/resources |
| 本地app/dmg打包与签名 | pass，candidate | desktop/dist/darwin-arm64；ad-hoc签名，未notarize |

native自动化是实际NSAccessibility输入路径，待绘制帧由该窗口CapturePage处理；另有安全只读Inspection回读UI状态。未声称物理鼠标或解锁桌面验证。保留所有失败run及其source/frontend哈希；后来通过的run不覆盖旧失败。

## 原生性能

GOWORK=off、Apple M3 Pro上，有变化文本的1000项目录headless局部重绘基准68343ns/op、2462B/op、35allocs/op。列表只绘制220px viewport相关项目，不遍历绘制全部1000行。该基准只证明局部渲染成本，完整启动/RSS/长对话端到端性能仍待最终制品验收。

## 所需外部条件与后续步骤

1. 通过现AgentKit平台配置让知识服务可用，或提供同测试账户可用知识Runtime；当前客户端不能修复服务端缺失的Session scope secret，且用户禁止修改后端源码。
2. 提供专用飞书/钉钉/企微Bot配置的本机路径或环境变量名，保留真实已有账号；不向IM发送测试消息。
3. 解锁当前macOS桌面，重跑系统文件选择、导出保存、取消与artifact原生下载。
4. 若需要真实image写入验收，提供精确目标镜像和时间窗口，分别执行已实现的现OpenAPI动作；不构建后端镜像。
5. 完成上述live gates，再做最终app/dmg身份/启动/性能/完整功能复验，更新releaseReady；目前不准出。

本轮任务fixture按created回执精确清理：退出后运行fixture与独立Go task已删除，失败原生run的本轮5个created fixture也已删除；其它历史同名任务保留。专用Session与失败证据保留，不按标题批量删除。无commit/push/远端发布，无Runtime/Worker源码变更。

本地候选DMG：OpenMPA 0.1.0 arm64.dmg，6383997bytes，SHA256 `d7a9812721739d097204e8fa7c597c75ff92d44933acb46dee2b0efcf58cc37d`。app binary SHA256 `53762e5e3b70a6909a680ec0ab3816e9a35ba7a676366bb106511151ecf9b9cc`。该制品为本地检查候选，未满足全模块准出。

最后全流程run20261026T001149.639506Z已通过：原生账户/目录/Runtime/认证/资源与image读取、已有渠道目录/默认shared绑定表单、选会话/双轮WebView对话与pending清理、原生日程创建/Run/暂停/历史/删除、原生Space导航/中文富文本与Markdown保存/持久重开。退出前原生export SavePanel检测锁屏，state=blocked/code=blockedDesktopLocked。当前验收二进制与前端hash保存在native-bundle-20261026T001149.639506Z.json；安装候选hash见上文，验收程序与正式app的角色分别记录。

最新标准依赖Go全包与vet通过，nativeui/host race通过；Node全量75files382tests保持通过。24项nativeui回归覆盖目录过滤/输入/epoch/CAS/原token恢复/Once精度/取消/错误显示/虚拟行可达性；最终后续工作仍按上方外部条件推进。
