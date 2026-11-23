# DSH 内的 Bohrium CLI：公开 Git 适配包

分类：Deep。涉及新增对六个已发布平台包的外部依赖、公开索引契约和宿主默认安装路径。影响产品与投递范围的判断已由用户在前序任务中确定，本 spec 只记录已实现功能与验收边界，不引入新的产品决策。

## Goal

DSH 用户用一条普通 Git 安装命令，就能在目标 profile 内得到一个可直接调用的 `bohr`：不需要宿主批准任何构建或安装脚本，也不受同一 npm 前缀里旧全局同名 CLI 的遮蔽。

## Current evidence

- 基线 `origin/main` `824c15791d09d86c28ce1b05b167050a4ce8ac56`。
- 本仓之外的原型实测（本次改动前完成）：在默认安装路径上，直接依赖顶层 `@dptech-corp/bohr-cli` 会因它的 `postinstall` 被 pnpm 构建策略拦住；同一前缀里已有的旧全局同名包会因安装锚点解析掩盖新装的副本。改为依赖六个无生命周期脚本的平台分包后，CLI 0.2.2-rc.5 与桌面版 0.2.2-rc.2 都在无需批准、`pendingBuilds` 为空的情况下装成，并通过宿主自带 shell 运行。
- 六个平台分包在 registry 上均无 `scripts`；`bohr guide` 的 25 条指南及其引用文件都在二进制内部，无需外部业务 Skill。
- 社区目录（`awesome-dsh-plugin`）支持 `tree/main/<dir>` 形式的子目录条目；其 npm Discover 索引只收“子目录包名已发布且已发布包的 `repository` 指回同一 GitHub 仓”的条目，本方案两条都不满足，且不应为此造假映射。

## Requirements

- `dsh/bohr-cli/` 是独立 Git 适配包 `@dptech-corp/bohrium-dsh-bohr-cli`，版本跟仓库发布版本。只有 manifest、薄入口和 Cordis patch；无生命周期脚本、不复制 Skill 正文、不依赖顶层主包，也不依赖本仓以外的业务 Skill。
- 六个平台分包锁同一个精确版本（当前 `2.9.55`），按宿主 os/arch 只装一个；没有匹配分包时入口只告警，不让整个 profile 失败。
- 入口借助宿主自带 shell 工具暴露绝对路径 `DSH_BOHR_CLI`，并注册一个指向 `bohr guide` 的 bundled Skill；不注册工具，不改变宿主原有权限或审批路径。
- `skills/` 插件保持原语义，与本包可在同一 profile 共存。
- 自有 Pages 索引为两个插件各生成一条固定 commit 的 Git 条目（`path: skills`、`path: dsh/bohr-cli`），沿用现有工作流，不新增跨仓自动化，也不部署。
- CI 覆盖 packed 内容、patch 行、六 pin 同版精确、真实安装后的入口与二进制行为、两插件共存，以及官方索引 validator 校验。
- 文档给出 Git 安装命令、pin 升级规则、伴生程序按需另装，以及平台验证范围。

## Decisions

- 承继已实测结论：依赖平台分包，而不是为顶层主包做 Git 包装。
- 不发布新的 npm 包；CLI 主包的源码不在本仓，因此不给它加指向本仓的 `repository` 映射，也不把社区条目 YAML 放进本仓。
- 精确锁版本：profile 内的 CLI 只在公开仓 bump pin 时前进，不自动跟随 npm 上的最新版。

## Acceptance criteria

- `dsh/bohr-cli` 的 packed 产物只含入口、patch 和 manifest，没有 `node_modules`、安装脚本或旧安装器。
- 从真实提交的 Git 来源安装后，宿主内 `"$DSH_BOHR_CLI" version -o json` 与 `guide list -o json` 返回 pin 的版本与非空指南，且入口不注册任何工具。
- 同一 profile 同时启用 `skills/` 与本包时，两边的 Skill 都能按名加载；移除其中一个不带走另一个的贡献。
- 生成的索引经官方 validator 接受，两条 git 源的 commit 与 path 分别指向真实包含该目录的提交与子目录；错误输入失败。
- 开发机上 Linux amd64 宿主可安装并运行该平台分包；Windows 与其余未实跑的架构组合在文档中明确标注为未验证。

## Scope and non-goals

本仓新增公开适配包、索引第二条目、CI 接线和文档。不发 npm 包、不部署 Pages、不写入社区仓、不做跨仓自动化，也不提供 CLI 主包的 npm Discover 映射。社区单文件提案在对外投稿前只作为本仓之外的交付物存在。

## Verification

依赖安装、`npm pack`、`node --test` 与索引生成在任务独占的开发机目录执行；本机只做编辑、同步与 DSH 宿主验收，宿主验收使用与工作树内容一致的临时 Git fixture 提交。详见 `docs/plans/2026-11-23-dsh-bohr-cli-git-adapter.md`。
