# 更新记录

## 0.2.0

### 重要变更

- 17 个 Bohrium Skill 改为通过 `bohr` 执行业务，使用前需要安装并登录 Bohrium CLI。旧版附带的 Python API 客户端、示例请求文件和 API 冒烟脚本已移除；直接调用这些脚本的自动化需要迁移到对应的 `bohr` 命令。
- 整套安装器将官方 Skill 同步到 `~/.agents/skills` 和 `~/.claude/skills`，并在备份后退役旧的 `.codex/skills` 同名副本。详细迁移规则见[安装与包结构](docs/cli.md)。

### 新增与优化

- 新增 Gemini CLI 扩展入口及 `bohrium-setup`，支持引导安装和登录 `bohr`。
- 17 个 Skill 提供完整任务流程及本地参考文件，覆盖计算任务、数据集、论文、沙箱等场景。
- 改进整套安装器的版本核对、备份和失败恢复，备份原有 Skill 目录及其中的用户文件。

上一个已发布版本为 `0.1.1`。
