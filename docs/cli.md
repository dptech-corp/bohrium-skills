# Skill 安装与包结构

> **CLI 版 Skill 安装提示：** 本仓库当前 Skills 通过 `bohr` 执行业务。此文的 `bohrium-skills-cli` 只负责同步 Skill 文件，不安装 `bohr`，也不代用户登录。先安装 `@dptech-corp/bohr-cli@latest` 并完成 CLI 登录。
>
> npm 包和 GitHub Release 二进制内嵌发布时的 Skill 快照，旧版本不会随 GitHub 更新。`0.2.0` 包含当前 CLI 版 Skills；使用 `0.1.1` 等旧版本时，请升级或按 [README](../README.md#2-安装-skill) 从 GitHub 安装。

`bohrium-skills-cli` 管理本仓库内嵌的整套 Bohrium AI Skills，将 `zh/` 或 `en/` 下的完整包同步到本地 agent 目录，包括 `SKILL.md`、`references/` 和源包已有的 `agents/openai.yaml`。主文档直接提供任务流程；无需先通过 `bohr guide` 取回正文。

## 完整 Skill 包与内置指引

以论文检索为例，安装的目录包含：

```text
bohrium-paper-search/
├── SKILL.md                 # 完整任务流程与命令示例
├── references/commands.md   # 随包提供的详细参考
└── agents/openai.yaml       # CLI 已有的 Agent 界面元数据
```

Agent 先读取 `SKILL.md`，再按任务需要读取本地参考文件，不需要通过 `bohr guide` 再取回整套步骤。中文包完整翻译主文档与已有界面元数据；详细参考沿用 CLI 英文原文，保留本地链接及章节锚点。Sandbox 的源包没有 `agents/openai.yaml`，保持源包结构。

CLI 仍是执行依赖。所用二进制与本地快照不同或需要跨模块指引时，可核对已安装版本：

```bash
bohr --version
bohr guide list -o json
bohr guide read bohr-paper-search -o json
bohr paper --help
```

`guide read` 的 JSON 正文在 `data.content`，`data.name`、`data.version` 和 `data.path` 标识来源；先检查 `ok=true`。业务调用按具体命令帮助和实时价目执行。

`bohr guide` 读取本机内嵌指引；`bohr agents skills` 管理云端 Skill Hub；`bohr skill install` 安装本地完整 Skill 包。

---

## 整套安装器

`bohr skill install` 是 Bohrium CLI 自带的命令，用于将一个本地 Skill 目录或 ZIP 安装到指定 Agent 目录。安装 `bohr` 不会自动安装本仓库的 17 个 Skill；内置的 `bohr guide` 则无需另装。日常安装入口见 [README](../README.md#2-安装-skill)。下方独立安装器保留给整套同步和已有安装的维护场景。

### 源码安装（当前仓库内容）

克隆本仓库后，在仓库目录运行；需要 Go 1.23+：

```bash
go run ./cmd/bohrium-skills-cli install --lang zh
# 英文版改为 --lang en
```

此方式使用当前仓库内容，目标目录、备份和迁移规则见下方说明。

### npm 安装（已发布快照）

```bash
npm install -g bohrium-skills-cli
```

`0.2.0` 安装完成后自动运行 `bohrium-skills-cli install --lang zh --json`，同步完整的 CLI 版 Skills。旧版 `0.1.1` 快照不包含这些内容，并按旧版行为同步到 `~/.agents/skills`、`~/.claude/skills` 和 `~/.codex/skills`。

**0.2.0** 只写两个规范根：`~/.agents/skills`（Codex、DSH 共用）和 `~/.claude/skills`。规范目录完整安装后，旧 `.codex/skills` 中同名官方目录副本会先备份再退役，保留其他自定义 Skill。指向规范 `.agents` 的根别名保留；指向 Claude 或其他共享树的根别名明确报冲突，不沿链接删除共享内容。同名普通文件或指向普通文件的链接保留，不伪称已经备份迁移。

新备份目录带唯一后缀，同一秒内连续同步也不覆盖原备份。

**跳过安装后自动同步**（CI 或只想下载二进制时）：

```bash
BOHRIUM_SKILLS_CLI_NO_POSTINSTALL_SYNC=1 npm install -g bohrium-skills-cli
```

**覆盖 release 仓库地址**（测试本地打包产物时）：

```bash
BOHRIUM_SKILLS_CLI_RELEASE_REPO=https://github.com/OWNER/REPO npm install -g ./bohrium-skills-cli-0.2.0.tgz
```

### GitHub Release 二进制安装

手动下载对应平台的 release asset：

```bash
# macOS Apple Silicon
curl -L -o bohrium-skills-cli \
  https://github.com/dptech-corp/bohrium-skills/releases/download/v0.2.0/bohrium-skills-cli_darwin_arm64

chmod +x bohrium-skills-cli
mkdir -p ~/.local/bin
mv bohrium-skills-cli ~/.local/bin/bohrium-skills-cli
export PATH="$HOME/.local/bin:$PATH"

bohrium-skills-cli install
```

各平台二进制文件名：

| 平台 | 文件名 |
|------|--------|
| macOS Apple Silicon | `bohrium-skills-cli_darwin_arm64` |
| macOS Intel | `bohrium-skills-cli_darwin_amd64` |
| Linux x86_64 | `bohrium-skills-cli_linux_amd64` |
| Linux ARM64 | `bohrium-skills-cli_linux_arm64` |
| Windows x86_64 | `bohrium-skills-cli_windows_amd64.exe` |
| Windows ARM64 | `bohrium-skills-cli_windows_arm64.exe` |

手动二进制安装不会自动同步 skills，需要显式运行 `bohrium-skills-cli install`。

---

## 更新

```bash
bohrium-skills-cli update
```

- **npm 安装**：`update` 会先更新二进制，再用新版本内嵌的 skills 同步本地目录。
- **手动二进制安装**：`update` 不会自动替换当前二进制；它会提示新版本 Release 地址，并同步当前二进制内嵌的 skills。

---

## 常用命令

```bash
bohrium-skills-cli status
# 查看 CLI 版本、本地 skills 同步版本、语言、状态文件和目标目录安装情况

bohrium-skills-cli list
# 列出当前 CLI 内嵌的官方 skills（默认 zh 版本）

bohrium-skills-cli install --force
# 强制重新安装全部官方 skills；适合恢复被手动删除或改坏的 skill

bohrium-skills-cli install --lang en
# 安装英文版 skills

bohrium-skills-cli update --check --json
# 只检查 npm 上是否有新版 CLI，JSON 输出，不写入任何文件
```

---

## 同步机制

- 同步只管理官方 `bohrium-*` skills，不会触碰其他自定义 skills
- 替换同名 skill 前会自动备份到 `~/.config/bohrium-skills-cli/backups/<timestamp>/`
- 版本和内嵌 skill set 绑定，可通过 `status` 查看本地同步状态

---

## CLI 的好处

- 不用反复下载和手动替换目录
- CLI 版本和内嵌 skill set 绑定，排查问题时状态可追溯
- 替换前自动备份，误覆盖可恢复
- 不影响自定义 skills
- 一次同步覆盖 OpenAI Codex、Claude Code、agents 等常见 skill 目录
