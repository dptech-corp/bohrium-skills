# 外部 DSH 插件不限制 DSH 版本

分类：Standard，元数据及索引契约修正。用户已明确要求外部 Bohrium Skills 不按 DSH 版本限制安装。

Spec：`docs/specs/2026-11-05-dsh-plugin.md`。

- [x] 回归先断言包和公开索引不声明 engines.dsh，且没有 DSH peerDependencies 造成隐式版本限制；在开发机记录现有范围造成的失败。
- [x] 删除包的 DSH 版本要求及索引生成中的 engines 字段，更新说明；保留现有 Node 运行依赖声明和固定 Git commit。
- [x] 开发机执行相关 Node 检查，通过真实宿主兼容检查确认各版本号都不会因插件声明被拒；复用相同内容的完整加载验证。
- [ ] 提交修正，经 main PR 的 CI 合入并重新发布公开索引；读取线上 JSON 确认无 DSH 版本条件。

Self-review：仅改变版本准入元数据，不修改业务 Skill 和 DSH 本身。实际测试版本单独记录，不将未测试版本标注成测试通过。
