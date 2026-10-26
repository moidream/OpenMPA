# OpenMPA 客户端重新实施台账

公开版本已将验收账户、资源与回执标识替换为示例或占位符；实际对象仅保存在本机私有证据中。示例不是可用部署。
状态：实施中。以最新用户目标和[现有API客户端合同](2026-10-25-existing-api-client-spec.md)为准。

## 源码回退

S0 pass：本轮Runtime隔离worktree恢复48个tracked文件、保存87个untracked文件；Worker恢复15个tracked文件、保存27个untracked文件。两者git status空、git diff HEAD空，HEAD分别5a40dcdc和7840aeaa。完整备份与摘要在`.openmpa/reverted-backend/20261025T144353Z/receipt.json`；没有删除工作树，也没有提交、推送或由本轮改变云镜像。其它源码仓库用户改动保持。

## 当前接口基线

S1/S2/S3/S4 read-only audit：账户1234567890，Runtime Ready/currentVersion27，运行镜像dev-scheduled-sync-retry-5560a271-20261025214704。OpenAPI200包含114路径；Session列表200、15条；渠道capabilities200（飞书/钉钉/企业微信、账号scope、profile和private能力）；日程200、11条。没有session-bots/Human/Voice/desktop-memory路由。既有验收Session knowledge500、execution-config404需继续诊断，不能称知识或配置全通过。`/channels/feishu/accounts`不是既有列表路径，客户端应使用`/channels`；`/agents`404表示当前Agent记录缺失，不新增目录语义。

## 当前任务与证据

S0已完成。S1平台宿主可复用现有实现；S2会话/Bot接线改为已有Session/execution-config；S3多渠道专用fixture验收待；S4知识500需既有接口/provider诊断；S5最终包和原生性能待。旧native完整基线两轮对话/日程/页面/save/cancel已通过，但其代码包含撤销接口接线，必须改造后重新绑定最终构建验收，不能当最终交付。

标准准出使用OpenMPA npm lint/typecheck/test/build及Go test/race/vet、bindings generate、生产GOWORK=off build、真实.app/.dmg/设备验收。后端本地PG/full-check/新协议测试不再纳本目标。旧台账完整备份于`.openmpa/client-restart-backup/specs/`。

## 2026-10-26 原生布局补充（整体目标paused）

用户明确恢复业务sidebar、保留左快捷栏，并要求sidebar尽量原生。当前实现已把顶部标题、当前位置与操作图标、48px快捷栏和240px业务侧栏迁到MyGo ui；WebView仅保留内容区域。项目局部安装并读取emilkowalski/skills中的apple-design/emil-design-eng。详细边界、两轮自审及当前证据见[原生壳层复核](2026-10-26-native-shell-review.md)。

本地75文件382测试通过；原生NSBundle辅助功能点击链路、折叠/展开、Space选择及编辑保存重开通过。原生导出在macOS锁屏处blockedDesktopLocked，保持未完成；其它设置/渠道/日程/知识内容表单仍WebView，已分N1–N5迁移切片。不会把原生壳层或旧后端版本的验收当作全模块完成。

最近只读Runtime基线为Ready/version29/OpenAPI114路由；之前version27为历史。本次paused UI修改不调用云业务接口，知识500、渠道专用配置、全模块性能等旧待办仍保留。最终准出及云端测试继续暂停。

## 2026-10-26 恢复后的原生内容实施

用户已恢复整体开发授权。N1–N4已接入原生账户/Runtime选择/身份/资源及镜像设置、渠道、日程、知识与产物；N5新增原生assistant/Space输入。Scope/connectionEpoch、durable原request/ClientToken、原operation回读及原生文件窗口绑定均在客户端完成，不修改后端源码。

本地验证：全量Node75文件382测试通过；Go全包、nativeui/host/ipc/session及knowledge/artifact race通过；GOWORK=off原生与文件服务测试、Go vet、typecheck、lint与前端构建通过。后续小改按当前变更重新验证，不能复用旧制品hash作为最终证据。

实机N1 run20261025T231506.922143Z通过原生账号/目录/搜索/选择/确认/业务身份/资源/image读取及官方Tool关联；后续原生全流程仍有自动化失败，失败记录保留，不冒充已通过。macOS当前CGSSessionScreenIsLocked=true，文件保存/选择实机gate仍blocked。Go live run20261025T234936Z当前账户目录35个MPA Runtime、绑定/Session读写通过；turn1/2与旧task生命周期读取历史manifest，应标previously_verified，不是本轮新对话或调度执行证据。

最新云只读基线：Runtime32/Ready/OpenAPI114routes，外部image为dev-main-progress-bd2e340-20261026071145；本轮客户端未更新云image/env/resources。知识仍HTTP500，当前32实例日志确认MPA_SESSION_SCOPE_SECRET is not configured，并出现OpenViking相关调用；需已有平台环境配置就绪。execution-config404说明该Session默认配置不可编辑，客户端保留对话并明确显示缺失。专用三渠道测试凭据路径、可用知识Runtime与解锁信息已请求，尚未取得不能准出。
