# DSH plugin implementation plan

> For agentic workers: 由当前 agent 在独立工作区逐项实现、开发机验证和最终 diff 复核；当前授权包含提交、推送并向 main 创建 PR；合并与发布不在本次范围。

**Goal:** 为现有完整 Skills 增加一个可安装的 DSH bundle 和可发布索引。
**Architecture:** `skills/` 是独立无安装脚本的包，薄入口注册 DSH filesystem provider；Node 标准库生成 Git 固定 commit 索引，手动 Pages workflow 托管。
**Tech Stack:** ESM、node:test、DSH filesystem provider、Cordis registry、GitHub Pages。
**Spec:** `docs/specs/2026-11-05-dsh-plugin.md`

## Global constraints

- 测试、依赖安装和构建全部在独占开发机目录执行。
- 不改业务 Skill、不改 CLI、不改根安装器；一个插件含 18 个英文 Skills。
- 不发布、不打 tag、不更改线上配置，不写用户现用 profile。
- 保留错误与未验证状态；未通过不能降级为成功。

### Task 1: Bundle 的资源和生命周期

**Files:** 新建 `skills/package.json`、`skills/index.mjs`、`skills/cordis.patch.yml`、`skills/locale/en.json`、`skills/locale/zh.json`、`skills/icon.svg`、`scripts/dsh_plugin.test.mjs`。
**Interfaces:** 导出 `name`、`inject=['skills']`、`apply(ctx)`；消费 filesystem provider 的 `registerProvider` 生命周期，产出 bundled candidates。

- [x] 编写打包与真实 registry 生命周期测试。测试通过 `npm pack` 产物安装包后导入公开入口，断言 18 个名称、完整正文、参考资源、项目/用户优先级和插件释放行为。
- [x] 开发机执行 `node --test scripts/dsh_plugin.test.mjs`，记录缺少独立 manifest/入口的失败。
- [x] 新增独立包和薄入口，使用 `includeDefaultRoots:false, watch:false, providerName:'bohrium', bundledSkillDir:packageDir`，在 context effect 中释放 provider。
- [x] 开发机安装依赖并重跑同一边界测试，再通过隔离 DSH CLI profile 验证插件安装、配置加载与移除。

### Task 2: 索引、文档与发布准备

**Files:** 新建 `scripts/build_dsh_index.mjs`、`scripts/build_dsh_index.test.mjs`、`.github/workflows/dsh-index.yml`、`docs/dsh.md`；修改 `.github/workflows/check.yml`、`scripts/check_release_version.mjs`、其测试、README 中英文、CHANGELOG、`.gitignore`。
**Interfaces:** `buildIndex({commit,publishTimestamp})` 返回 format 1 Git 子目录索引；CLI `node scripts/build_dsh_index.mjs --commit <40hex> --timestamp <seconds> --output <file>`。

- [x] 索引测试给固定 SHA 和 Unix 秒，验证 source、版本、缺失输入/错误 SHA/错误时间戳被拒；先在开发机记录失败。
- [x] 实现索引生成并用官方 validator 验证；将子包版本加入既有发布版本核验。
- [x] CI 安装验证所需固定版本并运行 Node 测试；新增仅 workflow_dispatch 的 Pages 构建与发布，不在本任务触发。
- [x] README 指向完整 DSH 指南：本地或 Git 固定 commit 安装，开启实验索引，添加源，运行 setup，按原生管理功能更新、禁用和卸载；无默认索引承诺。
- [x] 验证 bundle/索引检查、既有 Node/Python/Go 适用检查、最终 diff；记录公开托管和真实业务任务是否实测。

## Self-review

覆盖 spec 中所有产物与可观察边界。没有独立实现代理、不可用技能、自动发布、额外提交或未决设计选择。测试以真实宿主公开入口为边界，已由用户确认的安装/加载/索引/生命周期范围覆盖。

## 验证结果

见 `docs/validation/2026-11-05-dsh-plugin.md`。最终打包产物和原生宿主使用自带 runtime 模块；仅开发验证依赖位于子包 devDependencies。
