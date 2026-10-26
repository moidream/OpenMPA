# 原生内容页实施契约

公开版本已将验收账户、资源与回执标识替换为示例或占位符；实际对象仅保存在本机私有证据中。示例不是可用部署。
状态：active。用户恢复整体开发授权；本契约继承 existing-api-client-spec 的客户端边界，按 N1–N5 逐片完成。不修改 Runtime/Worker 源码，不构建后端镜像。本地 app/dmg 最终准出仍要求完整真实业务验收。

## N1 账户与 Runtime

原生设置分 Account、Runtime、Authentication、Deployment 四个页签。MyGo 安全输入、勾选、按钮、系统确认；Go 直接复用 Platform/Connection/Deployment/RuntimeConfig。登录成功自动查询所选 Region/Project 的 MPA 目录；搜索和选择保留精确 Runtime ID，只有完整加载且匹配当前 scope 的成员可连接。目录使用现 ListRuntimes 分页和 MPA 标签过滤，不硬编码测试 Runtime。

账户/连接 epoch 或页面切换会失效旧异步操作、确认框和输入凭据。切换后通知 WebView 回读 Connection.Status，清理旧会话及刷新当前目录。凭据不进入 ShellState、日志或证据；Remember 复用现 Keychain 服务，恢复仅使用安全 profile 引用。

Runtime 资源支持 Description、CPU、Memory、Min/Max Instances、Concurrency。先 Read 获取版本和 fingerprint，确认精确资源及目标值，再用原 UpdateRuntime。镜像以当前官方 ToolId 为权威，Runtime image 和 Worker Tool image 各自独立更新，原 operation 回执可重开恢复，未知结果仅回读。实现不会生成或推送后端镜像；本轮自动验证只读取云资源。

## N2–N4 域内容

渠道原生目录、账号配置、启停、绑定回读、权限、诊断及 profile/私有用户操作继承 Channels 服务的能力与账户维度。日程目录、各 schedule 类型、创建/更新/执行/启停/历史/删除继承 Schedule 的 ClientToken、ExpectedVersion 和权威 count。知识导入/版本/搜索/内容/删除及产物保存继承 Knowledge/Artifact；长文本可继续 WebView。未完成整个域功能时不得仅因列表原生化称该域迁移完成。

## 验证与准出

N1 回归：目录成员/region/project/search/完整性、数值解析、安全输入清除、迟到结果失效、原生布局与页签、原操作未知回读。N2–N4 增加各域的真实操作和隔离验收。最后重新生成绑定、执行 Node/Go 合同和集成测试、原生运行、构建 app/dmg；只对本次验证的最终制品出具准出。

云知识接口失败或专用真实渠道材料缺失时记录 blocked，不使用假返回值。保持目标账户 1234567890；Runtime 由账户列表选择，验收使用 r-example-mpa-runtime。外部流程发布的版本与配置以即时读取为准，自动验收不覆盖。

## N5 原生新建交互与隔离复核

新建 assistant/Space 使用 MyGo Modal 和原生输入；会话复用既有 Session Create/Rename，Space 复用 Workspace。新建 assistant 在请求前保存原 SessionID 和 name，未知结果只恢复该 ID；不会创建第二个随机 ID。打开或取消输入框不产生资源。聊天区的创建入口通过固定 Host.RequestConversation 打开同一原生对话框。

原生目录增加按名称/ID搜索，目录使用有界虚拟列表。ShellState.connectionEpoch 是生成会话目录的连接版本；知识、产物和 profile 升级的 Session 成员必须同时匹配当前 Go 连接版本。仅检查目录成员而没有版本绑定不足以防止 Runtime 切换时使用旧目录。

页面导航保留未提交的普通表单，清除输入凭据、授权二维码和旧确认框；身份/Runtime切换清除全部作用域状态。创建、执行和知识导入意图在0600客户端文件中隔离保存，不在重启后自动提交。Once 计划编辑保留原有秒、纳秒和 DST offset；只修改名称时不得重新解释本地墙钟时间。

真实原生验收需要分别记录辅助功能输入、只读 UI 状态投影、WebView对话与原生文件对话框证据；只读 Inspect 不修改界面或业务状态。可见截图、头less点击与实机辅助功能分别列出，不能合并成物理鼠标证据。

会话配置、改名、删除使用固定 Host.RequestSessionAction，由当前连接版本的目录和选中 ID 决定对象。配置用原 ExecutionConfig/If-Match，不存在的404显示 Runtime defaults，并保留聊天。改名/删除回读现Session服务并向WebView发送固定更新事件；确认框最后渲染以覆盖其他原生Modal。
