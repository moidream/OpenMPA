# Runtime 定时任务管理

日期：2026-10-29。范围：OpenMPA 桌面端；复用现有 MPA Runtime API，不增加云平台 Action 或客户端调度器。

## 接口与页面依据

- `agentkit-mpa-agent/app/api/v1/scheduled_tasks.py`、`app/schemas/scheduled_task.py`：`/api/v1/esa-cron-tasks` 的列表、新建、更新、删除、立即执行、执行历史和投递目标接口。
- `arkclaw-team/docs/mpa-scheduled-tasks-top-spec.md`：同一组 Runtime 接口及任务用户隔离、六种结构化计划的契约。OpenMPA 通过已有原生连接访问 Runtime，不依赖 TOP 网关。
- Studio `frontend/src/cronjobs/MpaCronTasks.tsx`、`MpaTaskEditor.tsx`、`mpaSchedule.ts`：列表/日历、状态筛选、分页、任务操作、时区投影和编辑交互参考。

## 最终行为

- 页面展示当前 Runtime、当前请求用户的任务，包括同身份在聊天中创建的任务。保持原生层鉴权，不替换用户身份或读取其他用户的任务。
- 原生服务完成分页读取后才展示完整目录，保留服务端执行次数和成功率；缺少统计信息时不计算或伪造统计。异常分页不显示部分数据为完整结果。
- 列表包含名称、启停状态、执行周期、上次执行、下次执行、最近执行状态、执行 Agent 和操作；支持搜索、状态筛选和每页十条。
- 日历按设备时区展示预计执行安排，遵循任务 IANA 时区及夏令时；暂停任务不展示预计触发。复杂 Cron 只展示服务端下次执行时间，单日超过三项可展开。
- 新建/编辑支持 Once、Interval、Daily、Weekly、Monthly、Cron，默认每天 09:00；单次执行使用设备本地时间输入并转为带时区的绝对时间。校验时区、日期、时间、间隔和重复日期，处理中文输入法确认。
- 保存时保留原任务的随机延迟、执行超时、Interval anchor、飞书机器人、接收目标及线程配置；新建/复制使用新的任务身份。飞书切换机器人后清空旧目标并重新读取允许的目标。
- 编辑和启停沿用 expectedVersion；创建和立即执行沿用原生适配器的 clientToken 恢复规则，不自动重试结果未知的写请求；UI 防止重复提交。
- 立即执行显示服务端返回的受理记录；详情提供历史刷新和继续加载，服务端状态原样保留并本地化已知状态。立即执行受理不表示执行成功。
- 身份缺失、接口不支持、加载失败、空目录和搜索无匹配分别展示；身份缺失可以跳转到身份认证设置。切换 Runtime 清理任务和表单，忽略旧请求返回。
- 操作菜单使用原生 Popover，支持外部点击、Escape 和方向键，避免表格滚动容器裁切。共享 ModalDialog 补充最上层 Escape 处理，覆盖异步禁用焦点按钮后焦点落到 body 的情形。

## 验证与限制

- 7 个前端定向测试文件、32 个用例通过，包含适配分页、统计、配置保留、写请求标识恢复、Runtime 切换、表单校验、多飞书机器人、日历和已有本地化/弹窗用例。
- `go test ./internal/schedule ./locales` 通过；typecheck、lint、design:check、桌面前端生产构建通过。
- `scripts/verify-desktop-schedules.mjs` 覆盖九组浏览器流程与中英文截图，使用 `scripts/desktop-schedule-fixture.mjs` 的测试数据，不访问真实账号。
- 未在真实 Runtime 中新建、执行、删除任务；真实可见范围取决于 Runtime 接口版本和当前配置的用户身份。执行记录目前展示原生层已有的安全错误码及会话/调用标识，不引入未脱敏的后端自由文本。
