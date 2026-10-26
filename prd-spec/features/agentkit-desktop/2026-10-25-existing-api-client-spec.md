# OpenMPA 现有 OpenAPI 桌面客户端合同

公开版本已将验收账户、资源与回执标识替换为示例或占位符；实际对象仅保存在本机私有证据中。示例不是可用部署。
状态：当前实施正本。日期：2026-10-25。Change ID：openmpa-existing-api-client。按用户最新指令及 `$dev-loop` 生命周期执行；此前要求新增 Runtime/Worker 接口的 E5/E6/Human/Voice 方案失效。

## 1. 背景、目标与范围

基于当前仓库 React 页面和产品策略，用 MyGo 系统 WebView 与 Go 宿主交付 OpenMPA 桌面工作台。agentkit-mpa-agent 是已部署服务；它和相关 mpa-codex-worker 源码只读。客户端使用 AgentKit 控制面 OpenAPI 发现及管理 Runtime，使用平台返回的真实端点消费 Runtime 现有业务 OpenAPI。

必须交付对话、Session、多渠道管理、Runtime 设置、定时任务、知识库管理、页面保存与原生文件交互；不修改后端源码，不新增迁移/路由，不构建、推送或自动部署后端镜像。用户主动在设置页修改配置属于固定 OpenAPI 操作，与开发/发布后端源码分开；实测优先只读，写入限制为本轮记录的测试对象。测试账户1234567890，cn-beijing/default，Runtime `r-example-mpa-runtime`。

用户最新补充确认：Runtime配置必须包含Runtime image与所关联codex-worker Tool image的更新。保留现有Deployment两独立固定动作：UpdateRuntime只写选中Runtime的ArtifactUrl/ReleaseEnable；UpdateTool只写GetRuntime正式ToolId的ImageUrl。目标引用用户提供，不从后端源码构建镜像；更新前Get复核关联与指纹，unknown保留原operation且只回读，分别验证平台Ready、新版本及新沙箱实际采用。不能按Tool名称猜关联、不能以旧沙箱仍运行推断新镜像已生效。

一个 Runtime 管理多个独立 Session。桌面“助手/Bot”使用既有 Session 创建、重命名、执行配置和渠道 BotProfile 接口实现；不依赖不存在的 `/api/v1/session-bots`，不把 `/agents` 单个 Runtime Agent 当多个 Bot 目录。角色/skills 等只能在服务端现有支持的位置配置，Session execution-config 仅提交已支持 category。

原 Computer/Human/Voice/desktop-memory 扩展依赖未部署的新增协议，停止产品接线和后台请求；这些不构成本次用户列出的必交付模块。不使用禁用按钮或长期 unavailable 替代所列必交付模块。

## 2. 接口与数据权威

| 模块 | 现有接口/权威 | 客户端结果 |
|---|---|---|
| 云登录 | STS GetCallerIdentity；AgentKit `2025-10-30` 官方Go签名 | 核账号，AK/SK/STS仅留当前进程内存；不提供Keychain保存 |
| Runtime目录/切换 | ListRuntimes `veadk:agent-type=mpa` + GetRuntime | region/project过滤和名称/ID搜索；切换递增epoch，迟到事件拒绝 |
| Runtime配置 | GetRuntime/GetTool、现有UpdateRuntime/UpdateTool | 固定typed字段，回读关联Tool；不自动写、不复制秘密Envs；用户review目标后才提交 |
| 对话/Session | `/api/v1/sessions` CRUD、`/{id}/run`、POST SSE、events | 原Session权威；原幂等key/pending，历史恢复，不重复执行 |
| Session配置 | `/{id}/execution-config` GET/PATCH、If-Match | 真实revision/etag、model/mcpServers replace/clear/inherit；已accepted执行保持原配置 |
| 多渠道 | `/api/v1/channels/capabilities`、accounts、bindings、config、enable、diagnostics、permissions、profile | 飞书/钉钉/企业微信按服务端实际能力；凭据只入不出；unknown绑定先读，不重复注册 |
| 定时任务 | `/api/v1/esa-cron-tasks`及已有update/run/runs/delete | 服务端调度；创建、编辑、启停、立即执行、历史与删除完整闭环，退出桌面后仍调度 |
| 知识库 | `/api/v1/sessions/{id}/knowledge` import/list/read/search/content/delete | 文本/URL/原生选文件，异步状态，内容/引用与删除回读；无需新后端 |
| 页面/产物 | 本地Space/Page存储；既有Session files/download | 沿现有编辑器与审阅策略；CAS/草稿/重开/导出/取消，产物经Go原生保存 |

配置的对象/数组类型必须以现有schema和真实响应为准，不能以过往自拟DTO替代。例如 `effectiveRefs.model` 是带id的对象而非字符串。GET 404按实际缺配置/对象处理；控制面与业务面权限不互相推断。Session配置404不阻断既有对话读取，输入控件只显示实际支持项。

设置页Runtime配置第一片固定支持Description、CpuMilli、MemoryMb、MinInstance、MaxInstance、MaxConcurrency，字段由现有UpdateRuntime SDK确认。Go读取安全投影、记录当前version/字段指纹，用户显式提交前重新Get核对同目标同版本；只发送这些字段、RuntimeId和明确ReleaseEnable，不传ArtifactUrl/ToolId/Envs。该操作不会开发或更改镜像。JSON durable intent在提交前落盘，original operationId恢复只能Get回读，不自动再次Update；匹配字段、新version及Ready才显示配置生效，accepted不作完成。平台没有已证实CAS，保留外部GET→Update竞态说明。开发验证默认只读，真实配置写入需用户明确目标值。

### 顶部原生导航与 Runtime 目录补充

页面根容器第一个元素必须为浅色 macOS 标题栏，顶部坐标0，无额外上方内容或外边距。用户后续纠正：业务侧栏必须保留，不能把所有业务菜单、Session和Space挤到顶部。左到右为 MyGo HiddenInset 保留的 AppKit 原生红黄绿交通灯、OpenMPA 文字、分屏图标。保留窗口 frame，使用 traffic-light inset 留位与 drag/no-drag；不得用网页假圆点替代。分屏图标展开/折叠标题栏下方左侧240px业务侧栏，其内保留纵向业务菜单与 Session/Space 列表；折叠不改变选中会话、当前页面或草稿。右侧页面内容占据其下剩余空间，标题栏始终留在顶部，不参与内容滚动。真实 NSBundle 验收坐标、浅色、展开/折叠、原生控件与侧栏展开/折叠截图。

设置页登录后自动调用当前账户的 ListRuntimes，Go 以 `veadk:agent-type=mpa` 筛选支持的实例。支持 region/project 与名称/RuntimeId 搜索、手动刷新、空列表和失败提示；仅列表中实例可选择。账号、region/project 改变后旧列表即时不可选，迟到响应不得恢复旧目录；绑定时 Go GetRuntime 复核。选择更新全应用 Connection epoch，拒绝旧请求/事件，清空旧 Session 投影并重新读取，资源/镜像表单重置为所选实例。选中项明确显示并避免重复绑定；真实测试验证账户1234567890获取、筛选及选择r-example-mpa-runtime。不得以已部署实例一个硬编码选项代替账户目录。

顶部布局补充：`Your workspace / chat`（或当前 Runtime/Space）与侧栏展开折叠按钮放在同一44px顶部标题栏，垂直中心对齐。由左向右为原生交通灯、OpenMPA、侧栏按钮、当前位置、最右侧产物按钮；位置区弹性占宽，长名称省略并可悬停读取完整文本，两个按钮保持可见和可点。移除内容区独立contextbar，侧栏与右侧内容同时从标题栏下方44px开始。折叠侧栏不改变顶部元素的排列或垂直位置。此次用户指定的布局修改独立执行，整体开发目标保持paused，不触发云端验收或发布。

左侧快捷栏补充：保留48px快捷栏，位于窗口最左侧、标题栏下方44px开始，业务侧栏位于其右侧，内容区从x=288px开始。业务侧栏折叠时快捷栏仍保留、内容区从x=48px开始。快捷入口为首页、侧栏切换、新建会话、Spaces、定时任务和底部设置；图标按钮具有名称、提示、键盘焦点和当前状态。采用MyGo `ui`在原生窗口中绘制，与现有系统WebView组合；不得把CSS仿原生称为原生控件，也不声称MyGo绘制按钮是AppKit NSButton。顶部和业务侧栏也由MyGo原生ui管理，WebView只保留内容页面和IPC/文件/关闭协议，通过固定导航事件与宿主交互；原生渲染仅接受安全view/collapsed投影。快捷栏占位不创建重复网页按钮；固定Action枚举不会接受任意JavaScript或路径。

### 原生壳层与设计规范补充（2026-10-26）

采用用户提供的sidebar结构：原生助手列表、Spaces、最近会话、底部日程/知识/渠道/设置入口。顶部原生标题栏为窗口第一个区域，固定44px；其下左侧原生快捷栏48px、可折叠业务侧栏240px、右侧内容WebView共同布局。WebView DOM不再创建标题栏、icon-rail或业务aside。聊天和富文本编辑保持现有WebView实现；其它模块的内容表单暂保留现有实现，后续按可验证的独立切片迁移，不能宣称全部界面已原生化。

原生目录使用有界的安全Session/Space投影和虚拟列表；不在宿主复制凭据或原始消息。原生选择动作带精确ID，客户端验证目录成员并遵循既有草稿保护；ShellState版本拒绝迟到的旧投影。UI刷新只更新导航状态，不重发远端操作。原生按钮有系统键盘焦点、可访问名称和提示；选中背景与字体保持清晰对比。

项目局部安装用户指定`emilkowalski/skills`中的`apple-design`与`emil-design-eng`，对应`.trae/skills/`和`skills-lock.json`。应用系统字体、清晰信息分组、一致图标与间距、120ms轻量颜色反馈。高频导航和键盘动作不加装饰性位移动画；尊重系统ReduceMotion，不引入循环、重blur或全页面动画。本次只推进用户指定UI调整；整体开发仍paused，云端业务测试和最终准出暂停。

## 3. 安全、错误、恢复与性能

Go统一持有凭据、签名、网络、SSE解析、文件和持久状态；前端只调用固定typed IPC，无任意URL/Action/本机路径执行。运行不需要Node/Bun/Python sidecar。外链由系统浏览器打开，生产WebView限制导航与权限。

已接受请求恢复同invocation，网络或IPC结果未知时保留原operation/key并回读。只有既有服务端确证的幂等条件才允许同key重试；不以不存在的SessionBot接口验证durability。删除按精确ID，服务端逻辑删除不称物理清理完成。云镜像配置提交成功仅为accepted，Ready/配置生效/新沙箱采用与业务健康分别回读。

Session列表使用现有API，前端投影/渲染有界；历史避免每次事件重绘整页，无每Session自动探测不存在的接口。测量实际app+WebView启动可操作时间、RSS、包体、首事件投影、连续对话和富文本保存。用户未指定数字预算，报告实测值与具体回归；禁止框架hello数据冒充软件性能。最终发布依赖固定，`GOWORK=off`证明无需本地MyGo源码。

## 4. 渐进开发切片与验收

| 切片 | 实施文件/边界 | 必须证据 |
|---|---|---|
| S0 源码回退 | 两个本轮隔离后端worktree，备份后恢复；不动其他工作树/分支/云 | tracked/untracked备份摘要；git status干净、diff HEAD为空 |
| S1 平台/宿主 | desktop cloud/ipc/configstore/host、Settings | STS账户与MPA目录，Region/Runtime实际选择与切换，SQLite非敏感设置重开；固定Go签名/安全投影 |
| S2 Session/配置 | desktop session、services、BotSettings、Chat | 创建/重命名/删除、两轮对话/SSE恢复，现有execution-config GET/PATCH/CAS；无session-bots/Human请求 |
| S3 渠道 | channels、Channels/ChannelManagement | capabilities、现有账号读取；本轮专用注册/配置/启停/权限/诊断/解绑，精准scope且不发送IM |
| S4 日程/知识 | schedule/knowledge及页面 | 两模块CRUD与真实执行/入库/search/content；原生选文件，已创建对象回读及精确清理 |
| S5 原生页面/打包 | workspace、editor、MyGo配置与标准CLI | 原生编辑/保存/重开/export/cancel/close，实际.app/.dmg运行、性能/进程与包检查 |

每片按需求→影响分析→失败用例/当前行为证据→最小实现→定向验证→两轮review推进。稳定边界跑主仓库lint/typecheck/tests/build、Go test/race/vet及typed bindings再生成检查；只有OpenMPA为开发与验证对象，后端测试/部署门禁不再属于本任务。

本地fixture、PG、browser和真实云/原生设备证据分开。最终只有全部所列模块的实际运行证据才能准出；缺渠道测试材料或平台provider条件时保留明确未完成项，继续其它独立工作，不扩大后端开发。

## 5. 兼容、回退与交付

客户端数据schema不因删除未部署协议而删除用户草稿、既有pending/下载文件。旧原操作账本备份保留；不自动对未知远端请求执行补偿。代码回退只处理本轮拥有的两个后端候选，已有源仓库用户改动保持原样。

首阶段交付macOS arm64 MyGo `.app`/`.dmg`。Windows/Linux需各自实机，不假称已支持；公开签名/公证/更新源另按实际分发配置。提交、推送、合并与发布不从测试权限推断。准出报告包含源码/依赖/生成绑定/包hash、全模块证据、原生性能和未验证项。

## 2026-10-26 当前实施与准出状态

用户恢复后按N1–N5完成原生账户/Runtime/认证/资源/镜像表单、渠道/日程/知识/产物、assistant/Space新建及会话配置/改名/删除确认。细节正本为2026-10-26-native-content-spec.md。已经有完整新对话SSE/恢复、原生设置/日程/新建操作和客户端退出后Scheduled执行证据，仍因知识500、三渠道专用测试材料与锁屏原生文件验收未全部满足而不准出。现有app/dmg是本地candidate，无云镜像修改、无后端源码修改。具体证据以2026-10-26-native-content-acceptance.md为准。
