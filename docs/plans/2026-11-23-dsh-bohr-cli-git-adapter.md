# 公开 Git 适配包 dsh/bohr-cli

分类：Standard 落地、Deep 记录（新增外部平台包依赖与公开索引契约）。Spec：`docs/specs/2026-11-23-dsh-bohr-cli-git-adapter.md`。

- [x] 落地 `dsh/bohr-cli/{package.json,index.mjs,cordis.patch.yml}`，与已实测候选逐文件同哈希；`scripts/dsh_bohr_cli.test.mjs` 覆盖 packed 内容、patch 行、六 pin 同版精确和真实安装后的入口行为。
- [x] 在适配包测试里补两插件共存回归：同一个真实 Cordis registry 上同时挂载 `skills/` 与 `dsh/bohr-cli`，两边的 Skill 都按名可加载，移除一个不影响另一个。
- [x] `scripts/build_dsh_index.mjs` 增加第二条固定 commit 的 git 条目（`path: dsh/bohr-cli`），既有 `skills` 条目形状不变；版本漂移检查同时覆盖适配包。
- [x] `check.yml` 与 `dsh-index.yml` 安装 `dsh/bohr-cli` 依赖并运行新测试文件。
- [x] `docs/dsh.md`（中英）写清 Git 安装命令、pin 升级规则、伴生程序按需另装和平台验证范围；README 两份加一句指引；CHANGELOG 记录新增包。
- [x] 开发机独占目录：`npm install`、`npm pack`、`node --test`、真实工作树索引生成与官方 validator；记录真实结果。
- [x] 本机 DSH 宿主：用工作树内容的临时 Git fixture 真实提交走宿主默认安装路径 smoke，并验证两插件同 profile 共存；完全隔离并在结束后清理。
- [x] 本仓之外准备社区单文件提案，校验目录规则与 npm Discover 的范围差异；不放入本仓。
- [x] 自查最终 diff，清理临时产物，汇报未覆盖项与公开投递仍需的条件。

Self-review：只新增一个公开适配包及其索引、CI、文档接线，不改业务 Skill、CLI 与线上状态。本次变更按授权提交、推送到工作分支并开 PR，不合并、不发包、不部署 Pages、不对外提社区条目；社区条目落地需要公开仓先包含 `dsh/bohr-cli`。
