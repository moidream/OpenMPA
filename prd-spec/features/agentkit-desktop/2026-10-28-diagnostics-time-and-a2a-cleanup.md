# 渠道诊断时间格式兼容与 A2A 管理范围收敛

日期：2026-10-28

## 问题证据

用户在消息渠道页看到 invalidResponse 英文提示，机器人列表与群权限却能正常读取。只读查询 Studio 当前 Runtime 的健康诊断响应，确认 channel/appId 与所选机器人一致、configured/gatewayConfigured/routeConfigured 为 true、deliveryHealth 为 delivered；lastInboundAt 和 lastDeliveryAt 是带小数的 Unix 秒数。

Runtime 的 channel_control.observations 返回记录 updated_at 数字，桌面宿主原先把两个字段定义为 string，JSON 解码因此失败，整份健康数据被丢弃。回归测试先以数字时间戳复现失败，再加入兼容。

## 修复

- 宿主使用内部 RawMessage 接收可选时间字段，接受数字 Unix 秒、RFC3339 字符串、null/缺失，归一化为 UTC RFC3339Nano 字符串。对外 IPC 类型保持 string，不传播混合类型。
- 保留精确 channel/appId 校验；拒绝其他机器人的响应以及非法时间类型、越界时间，避免用宽松解析掩盖不可信数据。
- 渠道健康与权限读取错误按安全 code 本地化，不再直接显示 Go 的英文兜底。
- 不支持的模板目录只作为高级设置内的说明，不提供必定失败的读取按钮，不再触发顶部错误横幅。
- 按用户明确要求移除 A2A 的任务输入、会话使用按钮及相关提示，删除该页面的会话目标和草稿写入回调。保留空间绑定、成员注册/注销、成员详情、别名草稿与版本校验；未删除底层协作协议或修改远端资源。

## 验证

- pass：550 个前端测试，覆盖快捷入口不存在、别名草稿保留与保存、模板提示不触发请求/横幅、错误语言切换。
- pass：Go 全量 race 测试与 vet；覆盖数字秒、RFC3339、null、非法类型/范围以及其他 Bot 身份拒绝。
- pass：TypeScript、ESLint、格式、设计令牌、桌面前端构建与 diff 空白检查。
- pass：本地模拟页面确认任务输入及会话使用按钮均不再出现，截图只含模拟数据。
- pending：Go 后端修复触发开发版重启并清空内存登录，需要用户在本机解锁保存的凭据后再复查桌面真实诊断。

未创建、删除、解绑或更改真实机器人及 A2A 资源。
