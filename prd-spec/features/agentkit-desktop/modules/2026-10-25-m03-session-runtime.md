> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M03 会话、事件与执行控制 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3。首次设计，无 issue/MR。需求来源：[PRD](../2026-10-25-prd.md) FR-3/FR-4/FR-14、NFR-2/NFR-3/NFR-5。事实：证据（私有研究，不随版本发布） FACT-5–FACT-8。前置M02与可用测试Runtime。

## 1. 结果与适配范围

保留ThreadList、ChatTranscript、输入器和工具展示样式，替换CopilotKitProvider/useAgent/runAgent/connectAgent/stopAgent的运行路径。原生agentkit REST + POST SSE为唯一首版会话协议；AG-UI类型只可用于兼容展示，不依赖Intelligence/Thread存储或伪造AG-UI工具执行。

会话/历史属于选定火山Runtime中的MPA业务服务，客户端保存Space关联/选择/未决operation。通过M08绑定→M02业务API，不混平台Runtime、OpenDots对象、本地Page与A2A task。VeADK进程内虚拟session桥不是持久合同，不照搬合成创建/删除或错误空历史。

## 2. 会话生命周期

GET列表按当前连接主体，默认include_a2a=false。SessionBot以服务端sessionId为稳定核心，同Runtime允许多个；已有渠道session只有read/run/control分别有权才可查看/继续。管理渠道Key不授予聊天访问权；任意桌面Bot创建与绑定按E5，不给现有CreateSessionRequest偷加未消费的botId/appId字段。显式导入A2A会话需验证可见性与产品归属，不把系统任务会话全展示为用户聊天。

新建POST可用客户端canonical UUID建立本地未决关联；201后登记service sessionId与配置版本。超时重试创建同ID时需读取已存在对象并核验owner/绑定，不把409简单当成功。重命名PUT title，删除DELETE按tombstone回执移除当前可见入口，本地来源引用标记失效；删除不会声称立即清理debug数据或远程文件。

summary与events分开读，当前接口无分页，客户端不能虚构分页参数。历史数据过大时记录限制并列入后端演进，不静默截断“完整历史”。会话不可见返回404时不跨身份搜索恢复。

Bot选择不重绑当前session。已有绑定需服务端核验channel/appId、私有slot及pin快照；没有公开快照GET时修订标未核验，不用账号当前revision代替。E5创建返回BotRef与executionConfigVersion，再按本模块合同提交；仅Bot绑定存在不足以证明run持久幂等。

## 3. 幂等提交与版本

发送前创建客户端operationId、固定幂等键、contentDigest、确定executionConfigVersion和connectionScope。先持久化pending操作再请求run，保存正文仅用于恢复未决提交，成功后可移除pending正文。

所有新客户端run带Idempotency-Key，但自动同键恢复仅限已验证profile绑定的持久幂等路径；普通兼容路径不按键查原请求，header存在不是去重证明。profile绑定会话带确定executionConfigVersion。没有profile的普通入口可保留限制模式，提交结果未知只读核验，不自动重发；不绕过profile前提。

| 返回/场景 | 行为 |
|---|---|
| 接受回执 | 记录invocationId/turnId/operationId，再打开对应订阅 |
| 请求结果未知 | 显示pending；有持久幂等合同才恢复原键；普通兼容会话只读核验，无确证时用户决定是否重发并说明可能重复 |
| 409 session_busy | 不另建会话绕过；读取当前执行并保留输入 |
| 409 idempotency_mismatch | 停止重试；提示操作内容不一致，保留本地恢复证据 |
| 412 execution_config_changed | 显示服务端currentState；新版本需用户重新确认后新operation |
| 400缺前提/422 | 视为输入或兼容契约错误；不静默删除字段重发 |
| 404 | 会话不可见/已删除；停止提交 |

客户端busy仅防本窗口重复点击，不能替代服务端幂等与并发互斥。两窗口同会话的执行冲突由服务器裁决。

## 4. SSE投影契约

POST `/sse`带invocationId及lastEventId，不使用EventSource。解析跨chunk UTF-8、CRLF、多行data、注释、空frame和未知event；JSON失败保留安全诊断并启动历史核验，不能把半帧当消息。

建议ClientEvent字段为schemaVersion、connectionEpoch、subscriptionEpoch、sessionId、invocationId、eventId、kind、completeness和安全payload，属于新内部DTO。优先读取versioned customMetadata.mpa，支持已验证kind；无mpa或未知版本使用ADK安全fallback，不执行工具或链接中的命令。

| 事件 | 状态与展示 |
|---|---|
| message partial=true | 显示暂态文本/活动，绑定invocation与消息block，不当持久final |
| message持久final | 用稳定eventId及block关联替换对应partial/历史投影，不再次盲目拼接 |
| tool/file/plan/usage投影 | 展示安全结构；工具观察不等于可由前端执行或审批 |
| heartbeat | 更新连接存活；不改变游标/业务状态 |
| error terminal=false | 展示本轮告警/阻塞候选，保留订阅并核验control/history |
| error terminal=true | 更新本轮终态，仍可读权威summary/control |
| done匹配invocation且terminal=true | 标记对应运行终态；关闭本轮流并刷新summary，不影响其他运行 |
| done无terminal或不同invocation | 不宣布当前运行完成；按兼容版本读取权威状态 |

SSE没有id行，游标来自已完整消费的ADK event.id；heartbeats、done/error控制帧不得伪造游标。前端使用有界事件队列，文本批量合并与终态不可丢弃；如果消费失衡转入reconciling并从远端历史恢复。数值预算由OPEN-5采样确认，不靠无界缓存换流畅。

ADK函数调用、响应、thought和最终文本的映射由P0实际样本固化fixture。partial标识不足以单独证明一个字符串为delta或snapshot；只能按已验证字段语义追加/替换。无法关联的partial在final历史出现后整体重建对应invocation。Thought以安全摘要展示，不请求或保存未公开原始推理内容。

## 5. 重连与竞态

流中断首先标记reconnecting，远程状态保持lastKnown而非failed/cancelled。GET summary/events重建持久历史，再POST sse以当前游标与invocation继续；历史partial缺失时保留final，不凭空补齐。

410 cursor_expired清除该流游标，从完整可见events重建，再建立新订阅。历史与订阅交接会重放已知event，按稳定ID去重；若交接期间游标再次过期，保留已恢复文本、展示重连状态，进行可控只读恢复，不重复run。接口无法提供强原子快照时明确最终一致，终态后再读一次历史确认。

会话/连接切换递增epoch，旧事件不能更新新视图。执行控制操作固定taskId，不能对by-session/latest读到的新Turn误发旧按钮操作。客户端重启读取pending、summary/history/control，不以流是否关闭推测后端terminal。

## 6. 暂停、继续与取消

pause/resume仅对已导入且有权威A2A task绑定证据的任务使用Whole-Turn，按精确taskId读取/操作。普通REST聊天即使by-session返回latest记录也不自动启用：响应缺session/invocation/REST turnId关联，可能属于旧A2A任务。自动关联在M06-E1补齐。绑定核验后固定taskId/generation/allowedActions，请求包含expectedGeneration与固定幂等键。

desiredState或pausing表示请求已接受；实际paused及安全点需要回读，UI区分“正在暂停”和“已暂停”。409附当前control时更新状态，不自动换generation重发。503 durable store unavailable说明当前部署不支持该控制。resumeDisposition=new_turn_required提示“继续为新执行”，按允许状态调用continue并关联新Turn。

远程停止初期消费兼容`/abort`，按钮明确标为“停止此会话执行”，范围包含当前与排队执行；用户操作针对会话范围，不承诺仅停止所见Turn。aborted=false不显示成功，读取summary/history。刷新不能消除外部客户端在请求期间启动执行的竞态；需要精确Turn取消时先实施M06-E1并关闭兼容操作。A2A tasks/cancel需TIP及实际A2A task，不能将REST turnId代入，也未证明桌面用户owner授权，不作为兜底。禁止Whole-Turn pause失败自动abort。

Worker pause/resume/cancel与reconcile仅为诊断操作，需session:control和明确availability；不是Whole-Turn按钮自动兜底。调度暂停为enabled=false，只影响未来触发，M05单独处理。

## 7. 验收、验证与停止点

| 对应验收 | 必须场景 | 通过结果/证据 |
|---|---|---|
| AC-3 | CRUD、同ID创建超时、删除tombstone、重启继续 | 本地引用与远端summary/events相符；无跨owner恢复 |
| AC-4 | profile绑定同键重试、普通兼容提交超时、busy、412 | 幂等路径单次接受；普通路径不自动重发；没有自动新键执行 |
| AC-4 | SSE逐字节chunk、final重放、unknown event、error非终态 | 投影正确；终态不被连接错误覆盖 |
| AC-4 | 断网/跨pod/410/历史与流交接/晚到done | 文字去重；状态属于同invocation；无需新run |
| AC-4 | 已绑定A2A任务暂停中/409/503/continue；旧A2A记录+新REST run；abort false | 正确任务才开放control；旧记录不控制新run；无错误兜底 |
| AC-13/AC-21 | 同Runtime切换Bot、跨owner/slot、切换连接/会话、关闭窗口、重启pending | 旧事件不串写；远程任务继续可查 |
| AC-15 | 停用desktop runtime adapter | 可回旧Web路径；不修改远端既有历史 |

实现验证使用现有Vitest体系覆盖parser/view model，Go httptest控制chunk/错误/并发以及`go test -race`覆盖共享服务，再在真实MyGo与非生产Runtime完成会话/控制闭环。测试fixture来自脱敏协议样本，不能用样例执行代替真实运行。只在以上证据齐备后P2准出；本轮未运行产品测试。
