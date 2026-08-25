# Bohrium Skills

Bohrium 科研计算平台的 AI Agent 技能集合。

> 新接入用户推荐使用 [Bohrium CLI（`bohr`）](https://www.bohrium.com/bohr-cli/intro)。`bohr` 提供统一的平台命令，并内嵌与当前 CLI 版本匹配的 Agent Skills。本仓库继续保留独立的 `SKILL.md`，供已有集成使用。

[English](README_EN.md)

## 使用 Bohrium CLI

### 安装

```bash
npm install -g @dptech-corp/bohr-cli
```

支持 macOS、Linux 和 Windows 的 x64 / arm64 环境。安装后推荐先通过浏览器登录 Bohrium 账号：

```bash
bohr auth login    # 登录 Bohrium 账号
bohr doctor        # 检查本地配置与连接
bohr --help        # 查看全部命令
```

### 常用命令

先通过根命令查看 Bohrium CLI 提供的能力，再进入对应命令查看子命令和参数：

```bash
bohr --help
bohr wiki --help
bohr wiki search --help
```

其他能力同样按层级查看，例如：

```bash
bohr job --help
bohr sandbox --help
bohr lkm --help
bohr agents mentor --help
```

本仓库每个独立 Skill 对应的 `bohr` 命令见下方清单。具体子命令和参数始终以当前 CLI 的 `--help` 输出为准。

### Agent Skills

`bohr` 内嵌了适合 AI Agent 使用的命令指南，内容与 CLI 版本绑定。先查看当前版本已经覆盖的 Skills；未列出的命令仍可通过 `bohr <command> --help` 获取参数说明：

```bash
bohr skills list                              # 列出当前版本内嵌的 Skills
bohr skills read bohr-sandbox                 # 读取一个 Skill
bohr skills list bohr-sandbox/references      # 列出 Skill 的参考资料
```

升级 CLI 即可获得新版本命令及其配套 Skills：

```bash
bohr update
```

## 帮助与文档

- [Bohrium CLI 产品介绍](https://www.bohrium.com/bohr-cli/intro)
- [Bohrium CLI 安装文档](https://docs.bohrium.com/docs/bohrctl/install/)
- [Bohrium CLI 计费说明](https://docs.bohrium.com/docs/bohrctl/pricing)
- [Bohrium CLI npm 页面](https://www.npmjs.com/package/@dptech-corp/bohr-cli)
- [Bohrium 平台](https://www.bohrium.com/)
- 命令内帮助：`bohr --help`、`bohr <command> --help`

## 本仓库中的独立 Skills

以下 `SKILL.md` 继续保留，供已经依赖本仓库的 Agent 或自动化环境使用。新接入场景推荐使用 `bohr` 及其内嵌 Skills。

### 已有集成配置

> 本节仅适用于继续直接使用本仓库独立 Skills 的已有集成。新接入用户仍推荐安装 `bohr` 并运行 `bohr auth login`。

直接调用平台 API 的独立 Skills 需要 Bohrium AccessKey。可在 [Bohrium 用户设置](https://www.bohrium.com/settings/account) 中获取，然后通过环境变量提供：

![获取 AccessKey](docs/images/access-key-settings.png)

```bash
export BOHR_ACCESS_KEY="YOUR_ACCESS_KEY"
```

请勿将真实凭据写入代码或提交到仓库。

已有 Claude Code 插件用户可以继续使用原安装方式：

```text
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Skill 列表

| Skill | 对应 `bohr` 命令 | 说明 |
|-------|------------------|------|
| [bohrium-job](zh/bohrium-job/SKILL.md) | `bohr job` | 计算任务管理 — 提交、查询、终止、删除任务 |
| [bohrium-node](zh/bohrium-node/SKILL.md) | `bohr node` | 开发节点管理 — 创建、启停、删除容器/虚拟机 |
| [bohrium-dataset](zh/bohrium-dataset/SKILL.md) | `bohr dataset` | 数据集管理 — 创建、上传、下载、版本控制 |
| [bohrium-file](zh/bohrium-file/SKILL.md) | `bohr file` | 文件盘管理 — 列出、上传、下载、移动、复制、删除 personal/share 盘文件 |
| [bohrium-database](zh/bohrium-database/SKILL.md) | `bohr database` | 科学数据库 — 浏览领域数据库和表格数据 |
| [bohrium-image](zh/bohrium-image/SKILL.md) | `bohr image` | 容器镜像管理 — 查询、拉取、创建、删除镜像 |
| [bohrium-project](zh/bohrium-project/SKILL.md) | `bohr project` | 项目管理 — 创建项目、管理成员、设置额度 |
| [bohrium-knowledge-base](zh/bohrium-knowledge-base/SKILL.md) | `bohr kb` | 知识库管理 — 文献管理、标签、笔记、召回搜索 |
| [bohrium-paper-search](zh/bohrium-paper-search/SKILL.md) | `bohr paper` | 论文与专利搜索 — 关键词与语义检索 |
| [bohrium-pdf-parser](zh/bohrium-pdf-parser/SKILL.md) | `bohr pdf` | PDF 解析 — 提取文本、表格、图表、公式 |
| [bohrium-scholar-search](zh/bohrium-scholar-search/SKILL.md) | `bohr scholar` | 学者搜索与画像 — 检索学者及其科研成果与指标 |
| [bohrium-sciencepedia](zh/bohrium-sciencepedia/SKILL.md) | `bohr wiki` | 科学百科 — 搜索词条、课程、知识点和知识图谱 |
| [bohrium-tools](zh/bohrium-tools/SKILL.md) | `bohr tools` | 科学工具库 — 浏览、检索并查看工具详情 |
| [bohrium-web-search](zh/bohrium-web-search/SKILL.md) | `bohr search` | 网页搜索 — 开放互联网检索 |
| [bohrium-sandbox](zh/bohrium-sandbox/SKILL.md) | `bohr sandbox` | 云沙箱 — 创建临时云 VM 并运行 Shell/Python |
| [bohrium-lkm](zh/bohrium-lkm/SKILL.md) | `bohr lkm` | 大知识模型 — 知识检索、推理链、论文知识图谱和本地 PDF 异步抽取 |
| [bohrium-mentor](zh/bohrium-mentor/SKILL.md) | `bohr agents mentor` | AI 科学小导师 — 基于文献检索的科学问答 |

## 计费说明

收费 Skill 按调用或机时扣账户余额，可在 [科研资产](https://www.bohrium.com/assets) 查看余额与账单。最新计费规则以 [Bohrium CLI 计费说明](https://docs.bohrium.com/docs/bohrctl/pricing) 为准：

| Skill | 是否收费 | 定价 | 计价单位 | 定价说明 |
|-------|:------:|------|------|------|
| bohrium-job | 收费 | 见定价页 | 元/小时 | 开机后收取机时费，价格见 Job 定价页 |
| bohrium-node | 收费 | 见定价页 | 元/小时 | 开机后收取机时费，价格见 Node 定价页 |
| bohrium-sandbox | 收费 | 见定价页 | 元/小时 | 开机后收取机时费，价格见 Node 定价页 |
| bohrium-paper-search | 收费 | 0.05 元/次起 | 元/次 | 论文：普通版(type 0) 0.05、加强版(type 1) 0.1 元/次；专利：type 0 0.1、type 1 0.3、type 2 0.5 元/次 |
| bohrium-pdf-parser | 收费 | 0.05 元/页 | 元/页 | 触发 PDF 解析时收取，查询结果免费 |
| bohrium-lkm | 收费 | 检索 0.05 元/次；Parse 1 元 / cache 命中 0.1 元 | 元/次 | 检索类（search、reasoning/search、papers/graph、claims/{id}/reasoning、variables/batch）：每月前 1000 次免费，之后 0.05 元/次。Parse：1 元/次，cache 命中 0.1 元/次。feedback 免费 |
| bohrium-mentor | 收费 | 2.0 元/次 | 元/次 | 创建会话时收取，2 元/次 或 200 光子/次 |
| bohrium-sciencepedia | 收费 | 与 Tools 共享每月前 1000 次免费，之后 0.01 元/次起 | 元/次 | article、keyword、knowledge_graph 各 0.02 元/次；search/universal、get_wiki_index 各 0.01 元/次 |
| bohrium-tools | 收费 | 与 SciencePedia 共享每月前 1000 次免费，之后 0.01 元/次起 | 元/次 | search/hybrid 0.01 元/次；detail 0.02 元/次 |
| bohrium-dataset / bohrium-image / bohrium-project / bohrium-knowledge-base / bohrium-scholar-search / bohrium-web-search | 免费 | - | - | - |

## 目录结构

```text
bohrium-skills/
├── zh/              # 中文 Skills
├── en/              # English Skills
├── docs/            # 历史文档与资源
├── README.md        # 中文首页（本文件）
└── README_EN.md     # English README
```

## SKILL.md 格式规范

每个 `SKILL.md` 至少包含：

```yaml
---
name: skill-name
description: "一行描述。Use when: ... NOT for: ..."
---
```

- **Frontmatter** — `name` + `description`（含使用场景和排除场景）；可选添加 `version`。
- **正文** — 功能说明、API 端点、参数表、返回字段、代码示例、错误处理。
- **代码示例** — 使用 Python `requests` 风格，不硬编码任何凭据。

## License

MIT
