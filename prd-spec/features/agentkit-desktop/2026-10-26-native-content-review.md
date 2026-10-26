# 原生内容迁移两轮复核

公开版本已将验收账户、资源与回执标识替换为示例或占位符；实际对象仅保存在本机私有证据中。示例不是可用部署。
范围：N1–N5客户端内容迁移，现有API、MyGo原生ui、系统WebView聊天/富文本。无后端源码/镜像构建写入；保持原脏工作区和备份。

## 第一轮：实现与接线

检查生产与GUI验收用同一Go域服务；生成绑定一致；安全输入不进入导航DTO；Runtime目录限制账户/region/project/MPA标签；Runtime与Worker Tool两独立操作和原receipt；渠道保留channel+appId、默认shared、provider-specific输入字段；日程ClientToken/CAS/六类型字段；Knowledge文件Picker与Artifact保存父窗口；独立原生新建输入及Scope隔离。发现Channels类型名不是Service、原生文件调用缺CallerWindow、explicitTest模式名、异步表单闭包及重复PermissionKey，均已修复。

## 第二轮：恢复与边界

检查scope/epoch变化、关闭、迟到回调、页面导航、确认取消、原binding/二维码、任务/导入/assistant未知结果持久化、原请求不变、标准MyGo依赖和列表虚拟化。补充ShellState.connectionEpoch避免旧Session目录用于新Runtime；导航清凭据但保留普通草稿，身份变化清scope；任务创建/触发先持久化再请求，重启只显式恢复原token；知识导入正文与file request按Session隔离；assistant先固定ID后创建/重读。补充Once秒/纳秒/DST原offset保留，取消确认清action，按钮当前busy复核。

发现测试自动化在native目录重绘与异步busy期间可能命中旧辅助功能元素：验收改为等待实际服务UI idle并读取安全只读Inspection，真实输入仍通过NSAccessibility操作。失败run保留，未借Inspect直接修改UI或绕过API。新建任务成功但原生手动执行自动化尚未完全通过，不能作为全生命周期通过。失败任务仅按本轮created回执暂停；不按名字批量删除。

## 验证证据

- pass：Node75files382tests、typecheck/lint/frontend；Go全包和GOWORK=off全包、vet；nativeui/host/ipc/session及knowledge/artifact race。
- pass：N1真实账号目录/过滤/选择/连接/身份/资源/image读取，run20261025T231506.922143Z；后续新增逻辑由本地合同与新验收覆盖，旧hash不当最终hash。
- pass：Go live新Session696c7587-ac39-4991-8a9d-a2f8c412de54双轮SSE/恢复/terminal/history、一次性Session创建删除，run20261025T235302Z。
- pass：Go live独立task0864b968-eeee-42f9-a483-d7d4d0349bd7创建/CAS冲突/手动运行成功/历史/启停/删除，run20261025T235430Z。
- pass：原生任务新建经MyGo输入→确认→Schedule.Create→官方TaskID；失败后已按精确回执暂停。
- fail：完整原生GUI链路的后续手动运行/控件查找在部分run未通过，仍修复验证。
- blocked：knowledge现接口500，Runtime32实例日志缺MPA_SESSION_SCOPE_SECRET，涉及OpenViking；专用三渠道Bot材料缺；macOS锁屏文件选择/保存。
- candidate：标准MyGo打包app/dmg成功并ad-hoc签名通过；未准出。

仍需完整native run、客户端退出后Scheduled触发、知识文本/文件/版本/检索/删除、专用三渠道绑定/配置/权限/诊断、原生保存取消和最终制品复验。未通过项保留，不标目标完成。

## 后续实机补充

原生run20261025T235817.380127Z已通过完整日程创建→手动Run→原生确认→历史→暂停→编辑→启用。实际NSAccessibility控件输入由验收CapturePage处理该原生窗口待绘制帧后可稳定命中；检查动作与截图来自同一窗口。保留前面失败，不替换旧结果。精确fixture <private-fixture-id> 已设Once2026-10-26T08:01:00+08:00后关闭客户端，待服务器Scheduled触发验收。Go本轮独立定时任务执行和清理另有run20261025T235430Z。

客户端退出后的Scheduled触发也已通过：task a07a5527-c5ad-4a5f-bd73-0089b0ce8aff的run2b77018f-ad0f-4741-8c58-75a23ec03c58于2026-10-26T00:01:00Z调度，00:01:29Z成功结束、duration19084ms，权威executionCount=2，两次成功。该精确fixture验收后已删除，失败run本轮5个created精确fixture也有回执清理；原有其它任务不按名称触碰。

全流程run20261025T235850.743675Z、20261026T000023.454249Z已通过原生设置/渠道列表/双轮WebView对话/pending清理/原生日程操作和精确删除，后续Space导航首次辅助功能匹配不继续子节点的问题已修正；page-only run20261026T000208.659334Z通过文档完整编辑保存重开，导出停在blockedDesktopLocked。完整native文件保存依然不能称通过。

有变化文本的1000项目录headless局部重绘基准为68343ns/op、2462B/op、35allocs/op（Apple M3 Pro、GOWORK=off）；仅是该局部原生目录性能，不是端到端UI延迟或全应用性能。

最后forms run20261026T000855.502099Z通过原生assistant创建与ID回读、当前目录epoch投影、改名保存回读、会话配置对话框、删除确认/服务404回读、产物面板开关、Space输入创建。修复action handler只在渲染后更新当前ShellState的问题：接收动作时回读当前Host安全投影；原生验收在输入动作完成后等待服务UI idle再回读官方资源，避免用旧响应判断失败。前述失败native助手ID为专用验收资源，保留证据；不按标题寻找或删除未知资源。

最终全流程run20261026T001149.639506Z所有已启用场景至导出前通过，导出blockedDesktopLocked。辅助功能驱动匹配父行角色时必须继续遍历子节点，不能在同名AXRow处提前返回；不再依赖同名父元素或仅截图判断输入。真实输入在前后等待UI服务idle以读取最终API结果，失败日志留存。原生输入对话框错误有独立反馈，避免动作触发在chat时安全错误藏在隐藏内容页。
