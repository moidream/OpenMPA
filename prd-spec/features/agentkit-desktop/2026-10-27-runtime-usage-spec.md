# Runtime 使用侧适配

状态：已实现并完成本地验证；云端验收未执行。类型：L2 功能适配。日期：2026-10-27。Change ID：desktop-runtime-usage。
承接 existing-api-client-spec 与 native-content-spec；不迁移云端资源创建过程。

## 依据、范围与场景

已验证事实：`desktop/internal/nativeui/channels.go` 已提供三种渠道的绑定入口；
`session/configuration.go` 已支持会话模型与 MCP 引用覆盖，但缺少工具明细。
参考仓库 agentkit-mpa-agent 的 `app/api/v1/studio_metadata.py` 提供带 revision
的 AGENTS.md；`collaborating_agents.py` 提供空间目录、绑定、成员和别名；
`app/a2a/app.py` 的 resource-topology 扩展公布已挂载 Skill Space。
这些接口由 Runtime 执行，桌面端只提供发现、配置及调用入口。

用户连接已有 Runtime 后，可查看 Skill 与 MCP、编辑指令、绑定已有 A2A 空间、
选择成员进入聊天草稿。未连接、未授权或旧 Runtime 必须显示明确失败，
不得伪造空列表、成功状态或自动写入默认值。
不包含创建 Runtime/Skill Space/A2A Agent、发布 Skill、修改其他仓库或部署。

## 需求、方案与验收映射

| 需求（来源均为本次用户要求） | 设计与状态契约                                                                                                                       | 验收                                                                                             |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| FR1 渠道绑定可用             | 复用原生渠道页面及现有鉴权、绑定状态机，回归飞书/钉钉/企微契约                                                                       | AC1：既有渠道测试通过；真实渠道写入需专用测试资源                                                |
| FR2 Skill 使用               | 从当前 Runtime Card 发现挂载空间，以平台签名只读动作分页读取 Skill 名称/描述；选择后写入当前会话草稿，由 Runtime 执行                | AC2：只发现 mounts 关联空间；分页及身份切换受控；草稿不自动发送                                  |
| FR3 MCP 使用                 | 展示当前会话 MCP 名称、描述、工具，复用现有版本化配置入口                                                                            | AC3：真实响应形状解码；旧接口不可用时不显示为已加载空目录                                        |
| FR4 AGENTS.md                | 原生文本编辑，读取 content/revision，POST 原 revision；Main Profile 如存在则区分其覆盖范围                                           | AC4：读取/保存/409冲突，失败保留草稿；不把 Main Profile 404 当初始化许可                         |
| FR5 A2A 使用                 | 分页空间与成员、绑定已有空间、CAS 别名；仅用已保存的唯一别名寻址，@请求置于当前草稿开头并保留已有正文，正常 Session 链路展示运行结果 | AC5：绑定/解绑/别名请求与 Runtime 契约一致；跨空间与过期结果不可应用；无别名时引导保存并禁用调用 |
| NFR1 凭据与隔离（原契约）    | Studio Key 只由 Go 当前连接注入固定 Studio 路径；拒绝重定向；连接/平台 epoch 与页面 generation 隔离                                  | AC6：伪造 header、路径逃逸、身份切换、认证拒绝测试                                               |
| NFR2 写入恢复（原契约）      | CAS 冲突提示重新读取；网络或成功体损坏标记结果未知，禁止自动重发；用户可读回核对                                                     | AC7：冲突/丢失响应/畸形响应测试                                                                  |

## 接口与失败边界

固定调用 `/api/v1/studio/agents-md`、`/api/v1/studio/collaborators/registry/spaces`、
`registry/agents`、`binding`、`aliases/{agent_id}`，以及会话 `mcp-servers`。
AGENTS.md 上限 40,000 Unicode 字符；A2A 目录使用服务端页码且保留总数。
平台调用只增加 `ListSkillsBySpaceId` 只读动作，按 Runtime 消费契约读取 Items 后在本地分页，
不把其他 Skill API 的 PageNumber/PageSize/TotalCount 字段混入该动作。
任意 URL、凭据、工具 schema 不进入渲染层。
401/403 保留权限错误，404/405 显示能力不可用，409/412 要求重新读取。
所有远端文本按纯文本渲染。执行超时复用 Controller 的 60 秒取消边界。

Main Profile 是参考仓库尚未发布的能力：若支持则读取并以其版本更新，
不假定 r- Runtime ID 等于 mpaInstanceId，不因 404 直接创建配置。
共享 AGENTS.md 与 Main Profile 的作用范围必须在页面说明，避免误改 Worker。

## 取舍、兼容与回滚

选择原生资源与协作页面，复用 Runtime 执行和现有聊天反馈；替代方案为嵌入 Studio，
但其 BFF/平台凭据边界与桌面不一致，故不采用。未来统一公开 API 后可替换只读 Skill 目录来源。
不新增存储迁移；旧 Runtime 原聊天、渠道功能继续可用。回滚仅撤销本次代码，
已确认的远端配置需用户使用保存前内容与最新 revision 再次修改，不能靠客户端回滚恢复。

## 实施顺序与所有权

接口契约先行，然后资源服务、协作服务与聊天草稿入口独立实现，最后组装原生 UI。
共享 connection、controller 由主 Agent 单写；locales 交给单独文件 Owner；
独立 service、原生页面和前端已委派并行实现。模型均继承主会话配置。
吞吐检查点：阻塞步骤为接口 DTO 和作用域契约；独立流为服务/页面/前端；
共享状态以文件所有权隔离；最小拆分为三个功能面加本地化，稳定批次后集中生成与验证。
保留原有 13 个脏文件中的既有改动（基线补丁保存于工作区外）。

## 验证、观察与风险

本地 Go 契约测试、nativeui harness、TS 类型检查、前端构建、Go test/race/vet；
两轮代码审查分别检查契约一致性和失败反例。成功由读回 revision 与目录状态判断，
错误仅保留安全码/HTTP 状态，不记录凭据或用户正文。无本次发布或生产观察窗口。
风险：云端版本可能落后参考源码，Skill API 权限可能缺失，真实渠道写入需要专用账号。
关闭方式为可用本地 Runtime/原生 harness 验证；未执行的云端验收如实记入证据。
无已确认的 GateGap；这是缺失功能适配。设计自查：无阻塞歧义，保持使用侧边界。

当前 Runtime 成员目录过滤 Agent Card，卡片明细显示不可用；不从注册显示名称猜测
真实寻址名称。Skill Description 是可选字段，缺失/null 以空描述展示。
AGENTS.md 成功回执允许相同 revision（无变化保存），但必须返回请求正文且不能倒退版本。
最终验证及未执行项见 `2026-10-27-usage-acceptance.md`。
