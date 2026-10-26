> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M04 Spaces、页面与文档聊天 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3，持久化文档及来源关联。首次设计，无 issue/MR。需求来源：[PRD](../2026-10-25-prd.md) FR-5/FR-6/FR-14、NFR-2/NFR-3/NFR-4。事实：证据（私有研究，不随版本发布） FACT-1/FACT-2/FACT-7/FACT-9。前置M01；页面聊天需M03；本地策略需OPEN-2确认。

## 1. 结果、范围与替代方案

延续当前Space库、嵌套页面、Tiptap/Markdown、自动保存及审阅后保存策略。建议Go本地文档服务成为首批Space/Page权威，取代desktop路径的OpenDots page API；远程agentkit只接收用户明确发送的页面文本，不假设可以访问本机文件。

这是用户待确认的推荐方案。若必须首版跨设备同步，先实施M06-E4服务端实体与权限/CAS，再让本模块消费远程PageService；不得把Knowledge版本或identity space_id替代Page/Space。

## 2. 本地数据契约

建议schema1保存storeRevision、spaces、pages、conversationLinks、reviewReceipts和recoverableDrafts。Page沿用id/spaceId/parentId/title/content/revision/createdAt/updatedAt，并将sourceThreadId改为带connectionScope、sessionId、invocationId、eventRef的来源引用；旧数据迁移时保留原始引用而不猜远端映射。

| 操作 | 输入与不变量 | 结果/错误 |
|---|---|---|
| 创建Space/Page | UUID由服务产生；有效Space；title必填；content为Markdown | revision从1开始，实际持久化成功后返回 |
| 更新Page | expectedRevision；title/content/parentId变更；parent同Space且不可成环 | CAS成功revision+1；旧版CONFLICT，保留草稿 |
| 移动Page | 目标Space/parent存在，子树一致迁移 | 单事务完成；来源会话不改变远程授权 |
| 删除Page | 当前revision，显式用户动作 | 删除/调整子页在同事务，来源引用留可恢复记录 |
| 删除Space | 非空时需清楚列出含页面及影响；不联动删远端会话 | 本地删除/导出选择，远端无副作用 |
| 保存结果 | 来源/草稿digest/目标/receiptId；更新已有页面时CAS | 同receipt同payload返回已有Page；不同payload冲突 |
| 恢复草稿 | pageId、baseRevision、draftContent、capturedAt | 不将草稿写成已保存Page；启动提示恢复 |

标题160、正文100000沿用当前编辑器JS string.length口径，明确按UTF-16 code units计数；Go用等价UTF-16长度校验，不能换成UTF-8字节或rune计数。emoji等非BMP字符占两个code units；边界测试需覆盖它们。如后续改为Unicode code points，必须同时变更两端并评审存量兼容。保存来源和草稿不赋予远端工具访问权。

初始Go内存状态+JSON快照单写者，事务锁中校验revision/层级/回执，以同目录临时文件写入、sync和原子替换。有效旧快照保留恢复副本；一组实体不拆成无事务保障的多文件。多窗口共享同Go服务；第二应用实例拒绝写同store或启用明确独占锁，不能仅凭内存mutex保护跨进程。

快照目录使用App.Path(PathUserData)，秘密不在其中。启动schema未知或损坏只读展示恢复选择，不能自动新建空数据库覆盖。大型数据快照性能预算需OPEN-5采样；发现瓶颈后单独迁移无cgo数据库，不把技术选择伪装成产品规模事实。

## 3. 编辑、保存与退出

复用PageAutosave的状态与冲突策略，把Patch调用换成类型服务；Page DTO从server模块挪到shared客户端契约，避免前端依赖node:sqlite模块。保存串行化，每次只提交自己确认的revision；保存过程中有新编辑不被旧响应覆盖。

状态saved/dirty/saving/error/conflict保持可见。失败暂停自动重试，保留draft；用户选择重试或最新版本。收到新revision且dirty时进入冲突，提供查看最新、另存当前草稿及重做修改，不自动合并正文。

页面聊天、切页和原生退出共用flush/恢复协议。关键草稿变化保留恢复副本；退出时异步flush成功或恢复副本持久化才能关闭。磁盘不可写时可导出到用户选定路径、继续编辑或明确放弃，不能把未保存变更当成功。

## 4. 页面聊天与审阅保存

页面聊天关联为 `(pageId, BotRef)`，BotRef固定connectionScope及service sessionId；同Page可关联多个Bot，同Bot可用于多个Page。默认选择当前Bot，不因新增/打开页面自动创建或重绑Bot。用户明确选择“新建Bot”时经E5服务端创建/绑定，单次串行预约并按同canonical ID恢复；普通兼容会话仍按M03标明默认配置。页面材料进入所选Bot历史的影响发送前可见，不能宣称每页自动拥有独立Bot记忆。

发送前flush，读取确定Page revision和digest，并以有边界标记的文本附在content中；UI显示“随本次消息发送：页面标题 / 版本”。该文本不作为system instruction、权限声明或文件挂载。正文超出当前API允许容量时明确拒绝或让用户选取段落，不静默截断。

Agent回复由用户选择保存为新页面或预览修改后更新现有页面；显示来源会话、目标Space、正文和修改范围。确认后本地保存receipt绑定来源、草稿digest及目标；重复点击只生成一份页面，修改草稿需新确认。更新基于当前revision，冲突不自动批准新内容。

首批审阅是客户端内容保存动作，不声称暂停后端工具执行。原`review_space_page` Copilot HITL工具需要M06-E3 challenge/decision协议才能实现远程等待与恢复；不能保留旧`respond()`调用并返回假的已保存工具结果。

## 5. 导入、迁移与同步边界

默认无自动数据迁移。P0若发现旧OpenDots数据，先只读清点和导出，确认Space/Page/来源ID映射；复制导入新存储，记录源标识和导入回执，旧文件保持原样。因本轮未发现或读取实际本地数据库，不能声称不存在存量。

未来同步由M06-E4定义服务端owner/space membership/revision；本地UUID、离线draft和remoteId分开。首次同步先上传用户确认的副本并回读，不把服务端数据覆盖本地未提交草稿。同步停止后本地导出与只读访问仍可用。

## 6. 验收与回退

| 对应验收 | 场景 | 可判定结果/证明 |
|---|---|---|
| AC-5 | 两窗口同revision、父子环/跨Space、磁盘写失败、进程重启 | CAS/层级拒绝明确；有效快照/草稿可恢复 |
| AC-5/AC-13 | dirty/saving/conflict下关闭、迟到保存响应 | 原生关闭协议保留草稿；无错误saved |
| AC-6 | 页面聊天flush失败/成功、重复打开会话 | 失败不发送；成功附确定版本；引用去重 |
| AC-6 | 保存同回复两次、修改后重审、冲突更新 | 回执幂等；新稿新确认；冲突不覆盖 |
| AC-14 | Tiptap/Markdown切换、粘贴、IME、超字符限制 | 原文不无声丢失；快捷键与字符口径一致 |
| AC-15 | 旧数据导入/损坏schema/回退 | 不覆盖旧数据；副本可导出；未来schema只读 |

Go存储用临时目录、可控写故障及并发测试；Vitest保留/适配现有autosave用例；实际MyGo测试输入、关闭与恢复。本轮只写spec。回退保存新快照副本，不删除用户文件；没有必要删除任何worktree。
