> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M06 后端扩展与完整产品体验 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3。首次设计，无 issue/MR。需求来源：[PRD](../2026-10-25-prd.md) FR-10，及完整FR-5–FR-9；事实：证据（私有研究，不随版本发布） FACT-9–FACT-12。用户已授权完整软件开发、必要spec调整及指定测试Runtime验证；隔离后端/Worker按实际部署源码实施。代码合并与正式分发另行处理。

## 1. 结论与拆分规则

完整保留当前产品策略需要一些agentkit尚未提供的客户端接口。每项扩展单独冻结后端契约，先实现权限和数据语义，再接桌面控件；不把函数调用、提示词或本地UI按钮伪装成后端支持。

本spec定义行为与准入，以下E1–E8是建议子模块编号，不是现有API路径或已部署能力。具体路由/schema需子模块启动时在agentkit本仓库形成可独立评审spec，遵守其治理、身份、版本和验证规则。

平台接入由首批M08实现，List/Get/版本发现不再是未知未来功能。平台KnowledgeId/MemoryId/SessionStoreId/ToolId只表明部署关联，不证明与MPA Business Session/OpenViking/Worker一一对应；E2/E6/E7先核验绑定/provider，有可复用平台合同才复用，不用云资源CRUD替代业务授权。

| 子模块 | 必需结果 | 依赖 | 可独立停止点 |
|---|---|---|---|
| E1 身份/能力与控制统一 | 同主体可见动作、协议与服务能力；任务身份适配；精确Turn控制 | M02/M03/P0部署事实 | 仍使用已验证旧身份和读取接口 |
| E2 文件与执行引用 | 用户附件上传后有scope受控的远端引用，run明确消费 | M03；E1身份 | 下载/Knowledge现有接口保持 |
| E3 人工审批 | 服务端挂起副作用、客户端决策、幂等恢复 | E1；M03/M04 | 本地审阅保存仍可用 |
| E4 Space/Page同步 | 文档实体、membership、revision/CAS、同步与冲突 | OPEN-2；M04 | 本地编辑/导出保持 |
| E5 桌面SessionBot公开管理 | 任意Session的服务端Bot绑定/快照读取、CAS升级、可信owner与scope | M02/M03/M05；复用E1身份；P2b前置 | P2a已有渠道Bot与普通默认会话保持 |
| E6 Memory管理 | 可授权检索/版本修改/删除和审计 | E1；既有memory plane | 只使用已有受控Agent工具 |
| E7 Computer检查与接管 | scope受控屏幕/浏览器/文件/终端与接管 | E1/E3；M03/M05 | 事件与已知文件下载保持 |
| E8 语音与额外渠道连续性 | 短期语音授权、同会话关联；另行确认多渠道合流 | E1/E3/E5；M03 | 文本会话与M09现有多渠道管理保持 |

## 2. E1 身份、能力与精确控制

新增只读握手建议返回协议版本、经服务端验证的主体scope、可用功能族、动作授权、限制和可用性来源。不得返回秘密、信任客户端自报actions或以OpenAPI路由存在推断已授权。功能族例如session/wholeTurnControl/scheduler/knowledge/attachments/approvals，实际枚举在子spec冻结。

任务身份接入需明确RuntimePrincipal到scheduler owner/space/mpa的服务端可信映射。legacy兼容保留，迁移先读再双路径对照，不让客户端拼claims。提供清楚的能力拒绝与版本错误，CORS若保留Web消费者也要正确处理ETag相关头。

普通REST run当前未证明建立A2A control记录。扩展先定义REST operation/turnId/invocationId与control资源的权威映射，响应包含session/invocation/turn/task/generation和scope；latest记录不代替关联证明。为普通未绑定profile的run补持久幂等合同后才能开放自动恢复。精确取消绑定Turn及generation/幂等键，不能直接开放A2A管理取消给桌面用户；会话abort保留兼容，暂停/继续保留safePoint/checkpoint/new_turn_required。

验收：同身份在session/schedule/knowledge看到一致scope；无权动作拒绝；旧generation不能取消新Turn；协议不兼容明确拒绝。先后端负向/兼容测试，再桌面AC-2/AC-4/AC-8/AC-10。

## 3. E2 文件传递与执行引用

现有Knowledge imports/file可作为知识检索材料，不能替代执行附件。新增附件契约应提供owner/session绑定的attachmentId、displayName/mediaType/bytes/digest/status；本机路径不能作为远端引用，客户端不得向worker提交任意宿主路径。

上传需要用户原生选择、服务端容量和类型校验、幂等完成回执、失败/中断清理以及过期删除。run只接受服务端确认READY且同scope的引用，明确是知识检索还是sandbox执行挂载。download/list/read若追加，均校验session和路径，不能只依赖难猜ID。

文件编辑若需要，使用expectedVersion/ETag和写入许可，冲突保留用户稿；路径穿越、软链接目标及并发读取语义需后端定义。客户端不得增加任意shell执行来填补文件API。

验收：跨session引用拒绝；上传中断不会留下可执行READY对象；相同幂等上传去重；run实际读到选定文件内容；版本冲突不覆盖。容量阈值来自后端契约再进入验收，不猜通用限制。

## 4. E3 审批挑战与回执

后台副作用执行必须先受服务端授权和挂起控制。建议challenge携带challengeId、owner、session/turn/toolCall关联、目标动作、安全参数摘要、payloadDigest、有效期、状态与版本。客户端approve/decline必须回传确定challenge及expectedVersion/digest/幂等键。

challenge变更、过期或被另一客户端决定时拒绝旧决策；同键同决策返回原回执，不执行第二次。用户拒绝不转为Agent自动确认；程序文本和页面内容不充当批准。审批后执行失败仍保留批准与实际执行两个状态，并可恢复查看。

页面保存分两类：纯本地结果审阅按M04；后端请求保存若要完整恢复原CopilotHITL策略，需服务端可挂起的明确工具合同与文档所有权，不假造工具响应。第三方连接写操作同样使用后端challenge。

验收：未审批工具没有副作用；过期/变更/跨用户批准被拒绝；重复决定至多一次执行；客户端退出后challenge仍可恢复；approved不等于execution completed。

## 5. E4 Space/Page与同步

服务端Space包含独立ID、owner/membership、名称和revision；Page包含Space/parent、Markdown、revision和来源引用。身份workspace_id/space_id保持原意，通过授权关系关联文档组织而非重用值。

读取和每个写入校验member允许动作，更新用CAS；层级成环/跨Space移动和删除语义明确。同步需有可分页变更游标/tombstone与过期恢复；离线草稿保留baseRevision，冲突不自动覆写。

迁移由本地副本显式上传、生成local→remote映射、回读比对，再选择服务端权威。双轨阶段保持离线导出；停止同步不会删除本地用户文档。多人编辑、分享和审计只有相应契约通过才声明支持。

验收：跨member拒绝、CAS冲突、父子移动、断线恢复、双向冲突、删除tombstone和回退副本；内容/层级/来源映射均可对照。

## 6. E5 桌面SessionBot管理的最小后端扩展

Runtime内多Bot、账号模板、渠道Session pin和角色实际执行已经存在（FACT-28–FACT-35），不重建Bot引擎。缺口是任意桌面Session的公开创建/绑定、可信用户接续和读取/升级快照；作为P2b前置，不能全部推迟到P5。下表是拟议行为合同，未实现，不声明新增路由已存在；实际API/schema在agentkit独立spec中冻结。

| 拟议操作 | 输入与权限 | 结果与原子语义 |
|---|---|---|
| 创建SessionBot | canonical sessionId、已授权主体、明确配置或合法模板引用、幂等键 | 服务端原子创建session归属+Bot映射+pin快照+执行配置；返回BotRef/sessionId/profileRevision/digest/executionConfigVersion；同键同输入恢复，冲突不留下可运行未绑定session |
| 读取/列出BotSession | 当前可信主体、Runtime scope、可选准确BotRef | 只列可见session；返回服务端绑定、pin快照与可用动作，不暴露秘密；不能借渠道管理权限跨owner查聊天 |
| 接续渠道Session | 合法登录主体与服务端验证的渠道owner/subject映射、已知sessionId | 分别验证read/run/control，私有slot generation及scope；不由客户端传ownerUserId赋权，不按同appId自动合并历史 |
| 修改独立Session配置 | BotRef、expectedConfigVersion、完整配置或明确inherit语义 | CAS形成新修订；配置改变不修改其他Bot/全局Agent；active invocation时拒绝，原Turn快照保持；清楚区分仅当前Session与账号模板 |
| 定版升级 | BotRef、目标模板revision/digest、expectedConfigVersion、幂等键 | 原子检查绑定/修订/无活动执行，更新pin及统一executionConfigVersion；run版本校验、请求hash、Turn快照锁定同一Bot指纹；冲突返回安全当前状态 |
| 本地归档BotSession | 精确BotRef、本地归档状态 | 仅隐藏/归档本地入口，不删除远端会话/记忆/知识 |
| 服务端删除BotSession | 精确BotRef、当前版本与允许动作 | 沿用Session tombstone的memory/knowledge访问撤销及保留期清理；不级联渠道账号、云Runtime、日程、其他Bot或账号级记忆；保留来源/回执 |

创建不依赖先注册飞书账号。需要把现有channel_session_accounts的账号映射推广为可支持desktop-origin Bot的绑定，避免假造channel/appId碰撞；Bot模板当前仅appId主键，跨provider和desktop模板必须有显式复合scope或稳定profileRef迁移，保留旧飞书记录及旧快照。实现选择先比较独立Bot绑定表与兼容扩展表，具体schema与迁移不能在桌面代码中臆造。

角色使用现有agentsMd及模型/技能/MCP选择；权限由Harness校验，不新增全能toolPolicy。继承既有null与[]语义，不能把空数组迁移成禁用。Bot pin与Session override的优先关系保持现有合同；现有Bot升级不会推进执行版本（FACT-36），选择DEC-11：E5原子更新pin和统一executionConfigVersion，创建/变更/升级都纳入版本；run的校验、request hash与接受快照使用同一配置指纹，不能再在版本校验后合入不同revision的Bot模板。任何写成功回读pin指纹与执行版本；binding存在本身不证明run已进入持久幂等路径。

桌面SessionBot默认只关联自身业务session；memory标明session/bot/owner可见性及provider，不自动复制其他Bot记忆。现有飞书私有Bot受控同主体召回保持兼容；普通desktop Bot共享记忆策略在E6另行冻结。多渠道共同进入同Bot需另行定义授权主体匹配、输入来源、输出目的、群隔离、并发和幂等，首版不隐式合流。

验收对应AC-20/AC-21：同Runtime新建两个桌面Bot，不需新增Runtime或IM账号；角色/技能/MCP分别实际生效；旧快照不漂移；升级与run竞态按统一版本拒绝旧配置，不能只靠回读；同键创建至多一份；跨owner/slot拒绝；冲突/运行中升级保持原配置；普通旧会话继续兼容。先agentkit协议/迁移/负控，再真实Runtime，再桌面消费。E5实现与本地两轮复核已通过，线上候选及双Bot实际执行仍待验证，不把本地结果记录为部署成功。

## 7. E6 Memory检索与管理

复用现有Business Session/OpenViking/Bot memory plane，并明确哪类允许桌面直接管理。建议分session/user/bot scope返回来源、版本、可编辑性、completeness与引用；内部system指令/秘密不暴露。

读/搜索可以先交付；写/删除必须有资源归属、版本和审计。现有bot named replace没有CAS输入，不直接包装为安全多客户端编辑。删除需说明索引、检索及保留周期，不能宣称立刻物理清除。页面与memory不同，不自动将所有页面送入记忆。

验收：scope越权拒绝；局部检索completeness准确；版本冲突保持原文；删除后检索/回读行为符合合同；操作有安全审计。

## 8. E7 Computer检查与人类接管

复用远程sandbox/browser执行环境，客户端不启动自己的OpenBot后端。为Computer面板设计session-bound短期访问授权、只读screen/activity/file与显式键鼠/终端动作许可。截图显示不等于允许接管，shell/file写权限分开。

接管需要lease/version和Agent safe point，防止人类与Agent同时键鼠输入；释放/过期恢复或继续为新Turn须明确。屏幕和终端输出按权限可見、脱敏和背压传输，不携带连接秘密，不在可调用Go窗口中加载任意远程页面。

终端若仅查看输出，不能出现可输入命令的控件；交互shell属于新增执行权限，必须由后端明确授权。文件list/edit仍复用E2路径与版本规则。

验收：未授权屏幕/键鼠拒绝；接管lease互斥/过期/断线可恢复；Agent暂停与恢复真生效；用户操作不接触宿主文件和浏览器会话。

## 9. E8 语音与渠道

语音需后端提供短期会话授权、输入/输出流、转写、ASR/TTS provider与文本Turn关联。实时发声与计算可并行，但执行使用M03同一会话和权限；不是在renderer配置模型/speech永久Key。

用户授予麦克风权限后发起；拒绝权限、断线、静音、结束通话和取消计算分开。通话结束保留转写回执，不猜发言即授权外部工具。飞书/Lark、钉钉和企微的既有配置归首批M09，不等待E8。E8仅处理语音及用户另行选择的跨渠道连续性；Slack没有当前已验证接口，需独立provider合同。

验收：短期token过期拒绝、麦克风拒绝无启动、文本/语音历史同session、停止通话不意外取消后端任务、工具审批一致、无长期秘密暴露。

## 10. 发布、回退与子模块完成规则

每个子模块先在agentkit形成协议/权限/schema/迁移spec，再实施后端，验证真实Runtime，最后桌面消费。必要schema迁移保留已发布版本，兼容旧客户端；用户已授权为完成既定软件目标按需完善spec，实施记录必须写明最新合同、隔离范围和验证结果。

子模块日志包含安全resource/operation/version和拒绝原因，不写内容/秘密。通过后再开启对应能力开关，失败只停该功能，已接受远程工作仍检查状态。所有E1–E8的完成均需AC-10对应的正向、越权、重复、断线及恢复证据；路线文件完整不表示扩展已经实现。
