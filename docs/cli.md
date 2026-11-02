# bohrium-skills-cli 详细指南

> **CLI 版 Skill 安装提示：** 本仓库当前 Skills 通过 `bohr` 执行业务。此文的 `bohrium-skills-cli` 只负责同步 Skill 文件，不安装 `bohr`，也不代用户登录。先安装 `@dptech-corp/bohr-cli@latest` 并完成 CLI 登录。
>
> npm 包和 GitHub Release 二进制内嵌发布时的 Skill 快照，旧版本不会随 GitHub 更新。获取本次 CLI 版内容，请使用根 README 的 GitHub 安装方式；整套安装可在当前仓库运行 `go run ./cmd/bohrium-skills-cli install --lang zh`（需 Go 1.23+）。下文的 npm/Release 安装方式仅获取已发布快照，不能保证包含本次迁移。

`bohrium-skills-cli` 管理本仓库内嵌的整套 Bohrium AI Skills，将 `zh/` 或 `en/` 下的完整包同步到本地 agent 目录，包括 `SKILL.md`、`references/` 和源包已有的 `agents/openai.yaml`。主文档直接提供任务流程；无需先通过 `bohr guide` 取回正文。

---

## 安装方式

### npm 安装（已发布快照）

```bash
npm install -g bohrium-skills-cli
```

安装完成后会自动运行 `bohrium-skills-cli install --lang zh --json`，将中文版 skills 同步到：

- `~/.agents/skills`
- `~/.claude/skills`

Codex 和 DSH 共用 `~/.agents/skills`。默认安装会在规范目录完整安装后，备份并迁移旧 `~/.codex/skills` 中同名官方副本，保留其他自定义 Skill，避免 Codex 同时发现两份。备份目录带唯一后缀，同一秒的连续更新也不会互相覆盖。

**跳过安装后自动同步**（CI 或只想下载二进制时）：

```bash
BOHRIUM_SKILLS_CLI_NO_POSTINSTALL_SYNC=1 npm install -g bohrium-skills-cli
```

**覆盖 release 仓库地址**（测试本地打包产物时）：

```bash
BOHRIUM_SKILLS_CLI_RELEASE_REPO=https://github.com/OWNER/REPO npm install -g ./bohrium-skills-cli-0.1.0.tgz
```

### GitHub Release 二进制安装

手动下载对应平台的 release asset：

```bash
# macOS Apple Silicon
curl -L -o bohrium-skills-cli \
  https://github.com/dptech-corp/bohrium-skills/releases/download/v0.1.0/bohrium-skills-cli_darwin_arm64

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
