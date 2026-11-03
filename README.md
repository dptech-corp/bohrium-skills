# Bohrium Skills

通过 [Bohrium CLI（`bohr`）](https://www.bohrium.com/bohr-cli/intro) 使用 Bohrium 科研计算平台的 Agent Skills。

[English](README_EN.md)

本仓库提供 17 组完整的 CLI 原生 Skill 包：主文档包含任务流程、命令示例、计费与恢复规则，配套 `references/` 和已有的 `agents/openai.yaml` 一并分发。来源锁定为正式发布的 `bohr 2.9.19`，见 [来源锁文件](cli-skills.lock.json)。

## 1. 安装并登录 Bohrium CLI

```bash
npm install -g @dptech-corp/bohr-cli@latest
bohr --version
bohr auth login
bohr doctor
```

支持 macOS、Linux 和 Windows 的 x64 / arm64。安装 Skill 与安装 `bohr` 是两个步骤；Skill 本身不包含 CLI。本次核验版本为 **2.9.19**，推荐使用最新正式版。

登录由 CLI 引导用户完成浏览器授权。服务端或无交互环境先运行 `bohr auth login --help` 查看当前登录方式；按环境既有方式向 CLI 提供凭据，不把密钥写进 Skill、仓库或对话。

## 2. 安装 Skill

`bohr skill install` 已包含在 Bohrium CLI 中，无需另装 Skill 安装器。选择下面一种方式：

### Claude Code 插件

安装英文版 Skills：

```text
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Agent Plugins / Codex marketplace

本仓库同时是 [Agent Plugins 1.0](https://agent-plugins.org/) 包（根目录 `plugin.json`，`skills/` 指向英文 Skills）。在 Codex 中可先加 marketplace 再安装：

```bash
codex plugin marketplace add dptech-corp/bohrium-skills
codex plugin add bohrium-skills@bohrium
```

也可 clone 后把 Codex 指到仓库根目录（含 `.agents/plugins/marketplace.json`）。

安装后请先跑 **bohrium-setup**（Codex/ChatGPT 的 onboarding skill）：让 agent 安装 `bohr`、完成登录并执行 `bohr doctor`，再使用其他 Bohrium Skill。

### Codex、Claude Code 或 DSH

获取本仓库，安装需要的 Skill，例如中文版论文检索：

```bash
git clone https://github.com/dptech-corp/bohrium-skills.git
cd bohrium-skills
bohr skill install ./zh/bohrium-paper-search --target codex --scope user
```

`--target` 可选 `codex`、`claude` 或 `dsh`；英文版将 `zh/` 改为 `en/`。同一个 Skill 选择一种语言，安装后按回执提示重新加载 Agent。`bohr` 内置的指引可直接通过 `bohr guide` 读取；让外部 Agent 发现本仓库的 Skills，仍需完成上面的安装。

安装后直接向 Agent 描述任务，例如：

> 帮我检索 2020 年以来锂金属电池 SEI 相关论文，整理标题、年份、DOI 和摘要；付费检索前先告诉我预计费用。

包结构、整套安装及旧安装器说明见[详细指南](docs/cli.md)。

## Skill 清单

公开 Skill 名称保留为 `bohrium-*`；CLI 内嵌指引使用 `bohr-*`，对应关系如下。

| Skill | 命令入口 | 内嵌指引 | 用途 |
|-------|----------|----------|------|
| [bohrium-setup](skills/bohrium-setup/SKILL.md) | `bohr` 安装 / `auth` / `doctor` | — | 安装 CLI、登录、体检，再交给具体 Skill |
| [bohrium-job](zh/bohrium-job/SKILL.md) | `bohr job` | `bohr-job` | 计算任务、日志与结果 |
| [bohrium-node](zh/bohrium-node/SKILL.md) | `bohr node` | `bohr-node` | 交互式开发节点 |
| [bohrium-dataset](zh/bohrium-dataset/SKILL.md) | `bohr dataset` | `bohr-dataset` | 数据集上传、版本查询与续传下载 |
| [bohrium-file](zh/bohrium-file/SKILL.md) | `bohr file` | `bohr-file` | personal/share 文件盘 |
| [bohrium-database](zh/bohrium-database/SKILL.md) | `bohr database` | `bohr-database` | 科学数据库与表结构 |
| [bohrium-image](zh/bohrium-image/SKILL.md) | `bohr image` | `bohr-image` | 容器镜像、构建与快照 |
| [bohrium-project](zh/bohrium-project/SKILL.md) | `bohr project` | `bohr-project` | 项目查询、创建、删除及成员查询 |
| [bohrium-knowledge-base](zh/bohrium-knowledge-base/SKILL.md) | `bohr kb` | `bohr-knowledge-base` | 知识库、文档上传与库内检索 |
| [bohrium-paper-search](zh/bohrium-paper-search/SKILL.md) | `bohr paper` | `bohr-paper-search` | 论文与专利检索 |
| [bohrium-pdf-parser](zh/bohrium-pdf-parser/SKILL.md) | `bohr pdf` | `bohr-pdf-parse` | PDF 解析与结果查询 |
| [bohrium-scholar-search](zh/bohrium-scholar-search/SKILL.md) | `bohr scholar` | `bohr-scholar-search` | 学者检索与画像 |
| [bohrium-sciencepedia](zh/bohrium-sciencepedia/SKILL.md) | `bohr wiki` | `bohr-wiki` | 科学词条、课程与图谱 |
| [bohrium-tools](zh/bohrium-tools/SKILL.md) | `bohr wiki tools` | `bohr-tools` | 科学工具检索与详情 |
| [bohrium-web-search](zh/bohrium-web-search/SKILL.md) | `bohr search` | `bohr-web-search` | 公开网页搜索 |
| [bohrium-sandbox](zh/bohrium-sandbox/SKILL.md) | `bohr sandbox` | `bohr-sandbox` | 云沙箱、执行与文件传输 |
| [bohrium-lkm](zh/bohrium-lkm/SKILL.md) | `bohr lkm` | `bohr-lkm` | 知识、论断、推理链与证据 |
| [bohrium-mentor](zh/bohrium-mentor/SKILL.md) | `bohr agents mentor` | `bohr-mentor` | 文献支持的科学问答 |

## 执行与计费

- 业务操作前读取安装包内的任务流程和相关本地参考，再用具体命令的 `--help` 确认参数。
- Skill 内的价目和例子属于锁定版本；调用前按指引使用 `bohr billing pricing` 核对当前费用、支付来源和用户授权，不凭缺失的扣费字段推算实际收费。
- `--yes` 仅用于已授权的具体操作；命令支持 `--dry-run` 时可先预览。自动化业务调用使用 `-o json`，同时检查退出码、`ok` 和业务状态。
- 异步请求已受理不代表操作成功；按指引验证终态和结果。CLI 缺少能力时报告缺口，不自动改为直接调用 API。

## 从旧 API 版迁移

原有的 17 个 Skill 名称、语言目录和 Claude 安装入口保留，内容已统一改为 CLI 版。旧版附带的 Python API 客户端、示例请求文件和 API 冒烟脚本已移除，`requirements.txt` 也随之退役。使用这些旧脚本的自动化需要改用 CLI，参数以当前帮助为准。

通过 GitHub 安装方式更新同名 Skill 后，按上面的安装和登录步骤准备 `bohr`。旧脚本可从 Git 历史追溯；历史 API 验证记录不作为当前 CLI 版的验收结果。[OpenAPI 规范](docs/api/README.md) 继续独立维护，本次不修改其接口定义或发布路径。

## 内容来源与验证

`cli-skills.lock.json` 记录来源 npm 包、正式版本、gitHead、公开名称与内嵌 guide 映射，以及完整文件清单。每种语言有 64 个文件：17 份主文档、31 份参考文档、16 份源包已有的 Agent 元数据。

英文主文档和参考文件取自同版本已发布 CLI。公开适配包括名称、运行依赖、代理占位符及逐项记录在锁文件 `patches` 中的修正与任务流程。参照 LKM，其他 16 个主文档补充按目标选路径、实际 ID 或结果的交接、完成标准和恢复步骤；这些内容按同版本实现核对，保留原生详细参考。已知修正包括 Wenyon 实例挂载、输出格式、Paper 翻页限制、Tools 分类与工具键的区分，以及 Web 上游结果容器。

更新时以正式 npm 产物为源导出选定 guide 的完整树，更新锁文件并复核中文正文；按本仓库既有检查与发布流程交付。CI 使用真实发布版逐文件对比英文包、核对完整参考与界面元数据，检查中英文代码示例、本地链接/锚点、关键流程边界，以及安装器在隔离目录中复制全部文件的结果。测试不调用业务 API，不能替代真实 Agent 业务验收。见 [.github/workflows/check.yml](.github/workflows/check.yml)。

## 帮助

- [Bohrium CLI 产品介绍](https://www.bohrium.com/bohr-cli/intro)
- [CLI 安装文档](https://docs.bohrium.com/docs/bohrctl/install/)
- [CLI 计费说明](https://docs.bohrium.com/docs/bohrctl/pricing)
- [CLI npm 包](https://www.npmjs.com/package/@dptech-corp/bohr-cli)

## License

MIT
