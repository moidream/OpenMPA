# 现有 OpenAPI 客户端方案复核

日期：2026-10-25。实施合同：[现有API客户端](2026-10-25-existing-api-client-spec.md)。按最新用户边界重新开始dev-loop。

第一轮检查：所有必交付模块映射既有控制面/业务面接口；对话与日程已有真实原生基线，知识接口在既有源码存在；不引入后端schema、迁移、路由或镜像构建。Runtime与Worker代码回退完成，干净HEAD分别为5a40dcdc、7840aeaa，备份见实施台账。SessionBot目录改用既有Session列表；Session配置按现有model/mcpServers category，不声明role/skills自定义已支持。

第二轮检查：云凭据与业务身份分槽，切Runtime拒迟到响应；原accepted invocation恢复与未知提交不重发；native下载/草稿/关闭均在客户端；镜像设置为明确用户操作，开发期不自动更新。新业务证据必须绑定当前Runtime版本和schema，历史旧镜像结果不作新验收。尚缺的实际渠道专用配置与知识provider响应需现场核验，不以修改后端解决。

结论：当前合同可进入客户端实现。采用dev-loop内建两轮自审；未宣称独立评审。必须整改客户端对未部署session-bots/Human/Voice/desktop-memory的接线，并核对真实execution-config对象类型。接口缺失/500属于实测待解决项，不证明知识库管理已正常。
