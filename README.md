# Bohrium Skills

通过 [Bohrium CLI（`bohr`）](https://www.bohrium.com/bohr-cli/intro) 使用 Bohrium 科研计算平台的 Agent Skills。

[English](README_EN.md)

本仓库提供 17 组中英文 Skill。每个 Skill 负责识别任务、检查 CLI，并读取已安装 CLI 的内嵌指引；命令、参数、计费和恢复规则随 CLI 版本更新。

## 1. 安装并登录 Bohrium CLI

```bash
npm install -g @dptech-corp/bohr-cli@latest
bohr --version
bohr auth login
bohr doctor
```

支持 macOS、Linux 和 Windows 的 x64 / arm64。安装 Skill 与安装 `bohr` 是两个步骤；Skill 本身不包含 CLI。本次核验版本为 **2.9.19**，推荐使用最新正式版。

登录由 CLI 引导用户完成浏览器授权。服务端或无交互环境先运行 `bohr auth login --help` 查看当前登录方式；按环境既有方式向 CLI 提供凭据，不把密钥写进 Skill、仓库或对话。

## 2. 安装本仓库的 Skills

### Claude Code

现有 marketplace 和插件名称保持不变，安装英文版 Skills：

```text
/plugin marketplace add dptech-corp/bohrium-skills
/plugin install bohrium-skills@bohrium
```

### Codex、Claude Code 或 DSH：安装单个 Skill

从 GitHub 获取当前内容，并选择 `zh/` 或 `en/` 中的目录。例如安装中文版论文检索：

```bash
git clone https://github.com/dptech-corp/bohrium-skills.git
cd bohrium-skills
bohr skill install ./zh/bohrium-paper-search --target codex --scope user
```

`--target` 可选 `codex`、`claude` 或 `dsh`。同一个 Skill 选择一种语言；其他客户端可将对应目录放到其支持的 Skill 目录。具体安装选项和同名冲突处理见 `bohr skill install --help`。

### 一次安装全部 17 个 Skill

具备 Go 1.23+ 时，可以在上面的仓库目录使用现有安装器源码：

```bash
go run ./cmd/bohrium-skills-cli install --lang zh
# 英文版改为 --lang en
```

这会同步到 `~/.agents/skills`、`~/.claude/skills` 和 `~/.codex/skills`，替换前自动备份。详见 [整套安装指南](docs/cli.md)。

**已发布的 `bohrium-skills-cli` npm 包和 Release 二进制内嵌发布时的 Skill 快照。** GitHub 内容更新不会改写旧发布物；当前 CLI 版内容请通过 GitHub 插件、`bohr skill install` 或本仓库源码安装。本次迁移不发布新的安装器版本。

## 3. 读取与 CLI 版本匹配的指引

```bash
bohr guide list -o json
bohr guide read bohr-paper-search -o json
bohr guide list bohr-paper-search/references -o json
bohr guide read bohr-paper-search references/commands.md -o json
bohr paper --help
```

Agent 使用 JSON 输出：先检查 `ok=true`，再读取 `data.content` 中的 Markdown 指引；`data.name`、`data.version` 和 `data.path` 用于核对来源，目录列表在 `data.items`。CLI 未指定格式时会按终端类型选择输出，非交互调用默认也是 JSON。

指引中的相对链接属于 CLI 内嵌资源，通过 `bohr guide read <guide> <path> -o json` 读取。它们不在本仓库的 Skill 目录中。遇到缺失命令或指引时先核对 `bohr --version`，并安装最新正式版。

`bohr guide` 读取本机内嵌指引；`bohr agents skills` 管理云端 Skill Hub；`bohr skill install` 安装本地 Skill 包。

## Skill 清单

公开 Skill 名称保留为 `bohrium-*`；CLI 内嵌指引使用 `bohr-*`，对应关系如下。

| Skill | 命令入口 | 内嵌指引 | 用途 |
|-------|----------|----------|------|
| [bohrium-job](zh/bohrium-job/SKILL.md) | `bohr job` | `bohr-job` | 计算任务、日志与结果 |
| [bohrium-node](zh/bohrium-node/SKILL.md) | `bohr node` | `bohr-node` | 交互式开发节点 |
| [bohrium-dataset](zh/bohrium-dataset/SKILL.md) | `bohr dataset` | `bohr-dataset` | 数据集、版本与文件 |
| [bohrium-file](zh/bohrium-file/SKILL.md) | `bohr file` | `bohr-file` | personal/share 文件盘 |
| [bohrium-database](zh/bohrium-database/SKILL.md) | `bohr database` | `bohr-database` | 科学数据库与表结构 |
| [bohrium-image](zh/bohrium-image/SKILL.md) | `bohr image` | `bohr-image` | 容器镜像、构建与快照 |
| [bohrium-project](zh/bohrium-project/SKILL.md) | `bohr project` | `bohr-project` | 项目、成员与额度 |
| [bohrium-knowledge-base](zh/bohrium-knowledge-base/SKILL.md) | `bohr kb` | `bohr-knowledge-base` | 知识库、文献与库内检索 |
| [bohrium-paper-search](zh/bohrium-paper-search/SKILL.md) | `bohr paper` | `bohr-paper-search` | 论文与专利检索 |
| [bohrium-pdf-parser](zh/bohrium-pdf-parser/SKILL.md) | `bohr pdf` | `bohr-pdf-parse` | PDF 解析与结果查询 |
| [bohrium-scholar-search](zh/bohrium-scholar-search/SKILL.md) | `bohr scholar` | `bohr-scholar-search` | 学者检索与画像 |
| [bohrium-sciencepedia](zh/bohrium-sciencepedia/SKILL.md) | `bohr wiki` | `bohr-wiki` | 科学词条、课程与图谱 |
| [bohrium-tools](zh/bohrium-tools/SKILL.md) | `bohr tools` | `bohr-tools` | 科学工具检索与详情 |
| [bohrium-web-search](zh/bohrium-web-search/SKILL.md) | `bohr search` | `bohr-web-search` | 公开网页搜索 |
| [bohrium-sandbox](zh/bohrium-sandbox/SKILL.md) | `bohr sandbox` | `bohr-sandbox` | 云沙箱、执行与文件传输 |
| [bohrium-lkm](zh/bohrium-lkm/SKILL.md) | `bohr lkm` | `bohr-lkm` | 知识、论断、推理链与证据 |
| [bohrium-mentor](zh/bohrium-mentor/SKILL.md) | `bohr agents mentor` | `bohr-mentor` | 文献支持的科学问答 |

## 执行与计费

- 业务操作前读取对应内嵌指引，再用具体命令的 `--help` 确认参数。
- 通过 `bohr billing pricing --help` 选择当前价目查询方式，按指引确认费用、支付来源和用户授权；本仓库不维护重复的定价表。
- `--yes` 仅用于已授权的具体操作；命令支持 `--dry-run` 时可先预览。自动化业务调用使用 `-o json`，同时检查退出码、`ok` 和业务状态。
- 异步请求已受理不代表操作成功；按指引验证终态和结果。CLI 缺少能力时报告缺口，不自动改为直接调用 API。

## 从旧 API 版迁移

原有的 17 个 Skill 名称、语言目录和 Claude 安装入口保留，内容已统一改为 CLI 版。旧版附带的 Python API 客户端、示例请求文件和 API 冒烟脚本已移除，`requirements.txt` 也随之退役。使用这些旧脚本的自动化需要改用 CLI，参数以当前帮助为准。

通过 GitHub 安装方式更新同名 Skill 后，按上面的安装和登录步骤准备 `bohr`。旧脚本可从 Git 历史追溯；历史 API 验证记录不作为当前 CLI 版的验收结果。[OpenAPI 规范](docs/api/README.md) 继续独立维护，本次不修改其接口定义或发布路径。

## 内容维护与验证

`zh/<name>/SKILL.md` 与 `en/<name>/SKILL.md` 各维护一个任务入口。详细参考文档在 `bohr` 中维护，本仓库只引用 guide 名称和资源路径。

GitHub CI 检查全部 Skill、Claude 清单、安装器和真实发布 CLI 的 guide/help 命令，不调用业务 API。测试命令与版本要求见 [.github/workflows/check.yml](.github/workflows/check.yml)。

## 帮助

- [Bohrium CLI 产品介绍](https://www.bohrium.com/bohr-cli/intro)
- [CLI 安装文档](https://docs.bohrium.com/docs/bohrctl/install/)
- [CLI 计费说明](https://docs.bohrium.com/docs/bohrctl/pricing)
- [CLI npm 包](https://www.npmjs.com/package/@dptech-corp/bohr-cli)

## License

MIT
