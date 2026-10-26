# OpenMPA 客户端架构

当前架构：React现有UI → typed MyGo IPC → Go平台/业务适配 → AgentKit控制面与选中Runtime业务OpenAPI。MyGo系统WebView负责界面，Go负责网络、签名、流式投影、原生文件和持久草稿；无Node/Bun/Python运行时sidecar。

云控制面管理当前账户资源，Runtime端点管理会话、渠道、日程和知识。现有Session代表独立对话/助手，配置使用既有execution-config及渠道profile；不新增session-bots协议。宿主凭据不进入renderer，epoch拒迟到数据，未知写结果保留原请求与回读。

后端源码只读、无需为客户端重新部署。原后端扩展协议撤销，具体接口与验收见[当前合同](2026-10-25-existing-api-client-spec.md)。
