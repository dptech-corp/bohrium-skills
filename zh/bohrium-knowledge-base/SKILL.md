---
name: bohrium-knowledge-base
description: "通过 bohr CLI 管理 Bohrium 知识库、文献、标签、笔记及库内检索。当用户需要知识库相关操作时使用；详细命令与执行边界读取已安装 CLI 的 bohr-knowledge-base 指引。"
---

# Bohrium 知识库

通过 `bohr kb` 完成任务，详细步骤以当前 CLI 内嵌的 `bohr-knowledge-base` 为准。

## 1. 检查 CLI

```bash
bohr --version
```

需要安装或升级时，按当前环境的权限和用户约定执行：

```bash
npm install -g @dptech-corp/bohr-cli@latest
```

本 Skill 面向最新正式版 CLI，本次核验版本为 **2.9.19**。

## 2. 读取与 CLI 版本匹配的指引

业务操作前先读取：

```bash
bohr guide read bohr-knowledge-base -o human
bohr guide list bohr-knowledge-base/references -o json
bohr guide read bohr-knowledge-base references/commands.md -o human
bohr kb --help
```

内嵌指引中的相对引用属于 CLI 内嵌资源。例如 `references/…` 要通过
`bohr guide read bohr-knowledge-base references/具体文件 -o human` 读取，而不是到本地 Skill 目录查找。
根据任务继续读取相关引用，并用具体子命令的 `--help` 确认参数。

## 3. 完成任务

区分知识库管理、库内检索和公开论文检索。先确定知识库与文献标识，再按内嵌指引上传、整理或检索；保留引用来源，删除前确认对象与影响范围。

- 需要访问账户时先运行 `bohr auth status -o json`；未登录则使用 `bohr auth login`，由用户完成浏览器授权。凭据交由 CLI 管理，不读取或回显凭据文件。
- 按内嵌指引核对当前价格、支付来源和写操作影响；只有用户已授权本次费用或具体目标的操作时才添加 `--yes`。命令支持 `--dry-run` 时可先预览；保持 CLI 的确认机制。
- 自动化业务命令使用 `-o json`，同时检查退出码、`ok`、错误详情和业务状态。异步请求受理后，按指引查询到可验证结果；缺失字段或未知状态如实报告。
- CLI 缺少指引或能力时，报告版本与缺失项并提供升级建议。操作保持在 CLI 支持范围内，不自行拼接 HTTP 请求或恢复旧 Python 客户端。
