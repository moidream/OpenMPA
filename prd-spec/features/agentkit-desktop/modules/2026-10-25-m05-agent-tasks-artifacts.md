> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M05 SessionBot、任务、知识与产物 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3。未接受草稿按单Runtime多Bot要求修订，无 issue/MR。来源：[PRD](../2026-10-25-prd.md) FR-7/FR-8/FR-9/FR-14、NFR-2/NFR-5。事实：业务证据（私有研究，不随版本发布） FACT-9–FACT-12及Bot/渠道证据（私有研究，不随版本发布） FACT-28–FACT-35。前置M02/M03；渠道配置M09；页面组织M04；任意桌面SessionBot创建需E5。

## 1. 结果与实际支持边界

Specialist成为一个Runtime管理的SessionBot入口，保留原卡片、对话、Scheduled & activity和结果侧栏。Runtime是部署/复用管理器，SessionBot是独立执行上下文，ChannelAccount是消息入口和凭据配置。多个Bot无需多个Runtime。运行时已有渠道绑定及独立Profile快照；桌面创建任意SessionBot仍需E5公开合同。

GET /agents只提供Runtime默认Agent，不是Bot目录。渠道管理权限不授予会话read/run/control。日程仍有legacy身份前提；文件公开API只有下载。每个能力独立准入，不能由聊天连接成功推断全部可用。

## 2. SessionBot模型与配置

建议稳定BotRef以RuntimeBinding scope + service sessionId为核心，包含服务端核验的owner/principalScope、可选channel/appId/gatewayBotId/privateSlotVersion、Bot profile revision/digest及executionConfigVersion。displayName/avatar是本地展示元信息，不生成权限或替换真实bot ID。当前channel账号可能管理多个Session，各自保存pin快照；账号Profile是默认模板，不是每个Session现有独立可编辑资源。

Specialist主列表展示有权SessionBot，按Runtime/渠道账号分组；账号设置另在M09。打开一个Bot进入其稳定session，切换Bot打开另一session，不把当前session重绑。已有会话仅在服务端允许时纳入列表；没有session读取权限不能用渠道管理Key或owner自报接管历史。服务端未返回pin快照时显示“修订未核验”，不把账号当前revision当成旧Session revision。纯普通REST会话标明使用Runtime默认配置。

独立配置已有agentsMd/model/skills/mcp：null在pin时继承Runtime默认；[]选择所有可用技能/MCP，绝不是禁用全部。没有独立role/toolPolicy/memory配置字段。模型、技能和MCP选择服从Harness已授权资源，不扩大工具权限。模板PUT使用expectedRevision完整替换并CAS；修改后新Session可采用新模板，旧Session保持原快照。

区分三种版本：账号Bot模板revision/digest、Session executionConfigVersion、Turn接受快照。当前Bot升级不原子推进executionConfigVersion，run版本校验/hash也不含Bot修订；因此只读执行版本不能证明完整Bot配置仍未变化（FACT-36）。BotProfile经effectiveRefs.botProfile进入Main/Worker实际执行；Session model/MCP override优先，agentsMd/skills来自BotProfile。通用execution-config仍只支持model/mcpServers；不能PATCH role类别来设置角色。当前Runtime全局profile-status/If-Match路径仅用于默认配置，不当Bot模板接口。

飞书现有Session profile-upgrade要求已绑定账号且无active invocation，升级到当时当前模板；没有公开expectedRevision/幂等请求合同，不包装成原子定版升级，不自动重发。通用Runtime profile-upgrade的If-Match/幂等不能移植到此路由。E5补快照GET及按目标修订的CAS升级，并将Bot快照变化原子纳入executionConfigVersion，接受/hash锁定同一配置指纹；通过后才用统一版本保护run。现有路径保留“提交前Bot配置可能被外部升级”的限制，不暗称已解决。

独立SessionBot不意味着独立容器/无限并发或默认完全禁止跨Bot召回。会话历史隔离来自service session；私有Bot记忆具有owner/appId/slotVersion/session维度，合法相同subject可受控跨Bot读取；scheduled scope收窄。没有Bot memory REST CRUD，Memories按E6管理。桌面新Bot的memory policy需E5/E6冻结，不能从头像或sessionId猜可见范围。

实现分两步：P2a消费有权已有BotSession与M09账号模板；P2b补E5后可从桌面创建独立BotSession。P2a只是可用增量，不能替代用户要求的完整首版SessionBot管理。首批不默认将多渠道合流同session；当前映射只绑定单一渠道账号，新增合流需要授权/路由/输出目标合同。

## 3. Scheduled & activity

后台任务由agentkit scheduler持续执行。当前任务schema未证明可指定任意桌面BotRef/profile修订；先显示真实task/session归属，Bot专属调度需E1/E5核验并补绑定合同，不能将本地分组当服务器Bot执行绑定。不存在`setInterval`/cron或系统tray保活替代服务器调度。

| UI动作 | 实际API与一致性 | 反馈 |
|---|---|---|
| 列表/详情 | `/esa-cron-tasks`分页、query/sort/overview | items/total/hasMore来自后端 |
| 创建 | POST name/prompt/schedule/clientToken等 | 以server taskId为权威，同clientToken恢复 |
| 改日程 | POST `/{id}` expectedVersion | 409重读，不自动覆盖新日程 |
| 暂停未来触发 | enabled=false的版本更新 | 明确“暂停日程”，当前运行不因此暂停 |
| 手动执行 | POST `/{id}/run` clientToken/mode Force | 不等于会话resume；恢复原clientToken结果 |
| 删除日程定义 | DELETE task，只删除定义与未来调度 | 不显示已取消当前run；已接受运行可能继续 |
| 查看/停止运行 | runs只给session/invocation；精确run取消等待E1 | 如另提供session范围停止，先核验scheduled session身份并单独说明范围 |
| 运行历史 | GET `/{id}/runs`分页 | 状态/次数/产物只用后端权威字段 |

Once/Interval/Daily/Weekly/Monthly/Cron和timezone按源码schema显示合法选项，不把旧intervalSeconds直接扩展为任意schedule。前端验证与后端错误同时保留。没有权威总次数时显示“权威次数不可用”，不根据分页记录数量估算。

key-auth/Runtime JWT聊天连接不能直接访问现有schedule身份路由。可选显式配置属于同授权用户的legacy任务凭据并核验scope，或者先实施M06-E1统一身份。禁止把本地Space ID填入JWT space_id、禁止从昵称拼userId。缺权限时任务页解释限制，保留后续接入位置。

VeADK task BFF仅注入Runtime Authorization和可信Studio owner，没有生成legacy X-Jwt-Token；不能用其代码证明任务身份缺口已解决。云IAM、RuntimeKey、Studio用户和MPA任务主体分别核验。

删除日程后保留本地task/run来源引用和已获得的安全回执，提供按已知有权session访问历史的入口；后端runs若不可再访问，明确显示定义已删除、历史当前不可用，不能伪造持续完整历史。未确认session可见性时不借用另一身份停止运行。

## 4. 产物下载与知识

产物候选来自版本化file投影、后端明确生成的可验证下载链接或用户明确输入sandbox path；它不是完整artifact registry，也不代表权限。相对URL只解析当前连接允许的download路由，拒绝模型提供任意host/凭据参数。缺可用path显示无法下载，不能猜本地或远端目录。

下载固定会话path接口，先校验归属，再原生Save对话框授予本地目标。Go流式写临时文件，Channel仅传进度；不经base64返回整文件。目标已存在时用户确认覆盖。取消/断流保留原目标，清理本次临时文件；上游没有可信长度/digest时标识“传输完成，完整性未由服务端校验”，不把EOF包装成已证明内容完整。

Knowledge接口可以将用户选定文本/文件导入服务端Business Session，使用幂等键及202状态轮询，明确显示知识用途、scope和版本。文件最大20MiB来自源码；索引尚未完成时不允许声称Agent已读文件。Knowledge不是上传sandbox附件；需要Agent执行读附件时转M06-E2。文件内容权限、选择与发送均由用户确认动作触发。

Memories页面暂不提供假的CRUD。受控MCP memory工具属于Agent工具面，只有M06-E6公开契约后才开放桌面直接管理。

## 5. 验收、观察与回退

| 对应验收 | 场景 | 结果与证明 |
|---|---|---|
| AC-7/AC-20 | 同Runtime双Bot、当前模板与旧Session、外部全局配置改变 | 独立pin配置实际消费；未核验修订不猜；普通默认会话明确受限 |
| AC-7/AC-21 | Bot模板CAS、Session升级运行中/结果未知、跨owner读取 | 草稿与旧快照保持；不盲重发、不借管理身份接管会话 |
| AC-8 | 同clientToken/版本冲突、关闭应用触发、删除定义后运行 | 单次创建/回读；删除不等于取消，运行/历史入口状态准确 |
| AC-8 | Runtime Key无task身份、缺权威次数 | 未授权清楚、无虚构scope/计数 |
| AC-9 | 有权下载/越权path/502/503/中断/覆盖目标 | 正确保存或明确失败；目标不被残缺文件替换 |
| AC-9/AC-10 | knowledge导入202/失败/版本/检索 | 按真实状态展示，不当sandbox上传完成 |

使用Go httptest +真实测试API回读验证输入、幂等和错误；使用实际MyGo原生对话框检查下载。日程触发验证需被授权非生产任务，清理任务使用其明确ID及任务范围授权，不触碰worktree。各能力独立开关，失败时保留M03聊天与M04页面；客户端回滚不撤销远程已经创建的日程。
