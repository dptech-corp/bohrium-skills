# DSH 接入验证记录

## 版本与范围

- 基线：`origin/main`，`688672f42acdc6c3582286610e6944318ff38acc`。
- 工作分支：`feat/dsh-plugin-20261105`；记录对应该分支的 DSH 接入变更。
- 开发机：`root@39.100.68.53`，任务独占目录 `/root/workspaces/bohrium-skills-dsh-20261105`；远端 checkout 的分支、HEAD、dirty 和同步状态已确认。
- 工具：Node `24.12.0`、npm `11.6.2`、pnpm `12.14.0`、Go `1.25.0`；原生宿主 DSH `0.2.1-rc.5`、Cordis `4.0.5`。
- Mac 仅进行编辑、同步和 diff 检查；依赖安装、测试和打包均在开发机执行。

## 自动检查

`DSH_TEST_PREFIX=/root/workspaces/bohrium-skills-dsh-20261105/ci-tools node --test scripts/dsh_plugin.test.mjs scripts/build_dsh_index.test.mjs scripts/check_release_version.test.mjs`：8 项通过，无跳过。CI 所需的最小依赖前缀也独立验证，未依赖完整 DSH 安装偶然提供的测试依赖。

- `npm pack` 真实产物可发现并完整加载 17 个业务 Skill 和 setup skill，正文及每个参考文件与来源一致。
- 真实 Cordis/DSH registry 验证 bundled 来源、资源基目录、项目/用户的同名覆盖、禁用、重新启用，以及移除后用户和项目内容保留。
- 官方 `@deepseek-ai/dsh-plugin-market-index@0.2.1-rc.5` validator 接受生成的 format-1 索引；移动 ref、错误 SHA、错误时间戳和缺少 CLI 参数会失败。
- 索引包含完整 Git SHA、`path: skills`、确切包版本和兼容范围；CLI 同时生成索引和图标。
- 发布版本检查可拒绝根包与 DSH 子包版本漂移。

已有 Python 单元回归：11 项通过。`PATH=/usr/local/go/bin:$PATH go test ./...` 全量通过。两份工作流和 bundle patch 经 YAML 解析；根安装器打包内容未改变。最终 `git diff --check` 通过。

## 真实 DSH CLI 与宿主启动

全部使用隔离的 `DSH_HOME=/root/workspaces/bohrium-skills-dsh-20261105/dsh-home`、`acceptance` profile。

1. 从本地目录安装，确认原生 profile 选择了 bundle；随后从实际 `.tgz` 产物安装，验证独立包路径。
2. 使用真实 DSH CLI 启动宿主。独立 probe 在宿主内通过 `ctx.skills` 发现 18 个 Skills，读取 setup、job、paper 正文和 job 的参考文件，并确认用户 sentinel Skill 保留。此阶段禁用 headless runner，不发模型请求。
3. 使用仅用于验证的 `0.2.1` 临时升级包替换 `0.2.0`：在 setup 正文加入测试标记，重启真实宿主后按名加载并断言标记出现。该测试版本未进入仓库，也未发布。
4. 使用真实 `dsh plugin ... remove` 移除包，确认组合配置中不再有插件行，用户 sentinel 文件保留。
5. 运行时使用 DSH 自带模块；最终独立包没有运行时安装依赖或安装脚本，原生安装不产生 peer dependency 缺失提示。

原始证据位于开发机任务目录：`real-dsh-boot.json`、`real-dsh-upgrade.json`、`real-dsh-boot-host.log`、`real-dsh-upgrade-host.log`、`dsh-install-host.log`、`dsh-upgrade-host.log`、`dsh-remove-final.log`、`removed-config-final.yml`、`node-tests-final.log`、`go-tests.log`。

## 尚未验证与交付状态

- 浏览器 Plugins/Discover 的实际渲染、图标显示和在线索引添加未做 UI 验收。
- 未调用 Bohrium 业务 API、未发真实模型请求、未执行付费任务；不能将安装/加载验证视为真实业务验收。
- 未发布 Pages 索引、未发布 npm、未打 tag，也未修改线上设置或用户现用 DSH profile。
- 开发机验证完成后，代码与验证记录一起提交至工作分支，向 main 创建 PR；验证结果不表示已经合并或发布。

## 公开索引与桌面版兼容修正

PR #39 合并后，已启用 GitHub Pages 并发布 `https://dptech-corp.github.io/bohrium-skills/index.json`。首轮发布 workflow `41996862066` 成功，公开索引直接返回 HTTP 200，源固定到 main 合并提交 `c4f00a7b72e4780c30b93bed5e233c112864bf43` 的 `skills` 子目录，无需 npm release 或新 tag。

桌面版 0.2.2-rc.1 的发现页已实际显示 Bohrium 条目与图标，但原来精确的 `engines.dsh: 0.2.1-rc.5` 导致安装按钮不可用。回归检查先复现对 0.2.2-rc.1 的错误拒绝，再将包与索引的共同兼容范围改为 `>=0.2.1-rc.5 <0.3.0-0`。上界排除 0.3.0 预览版，DSH 匹配规则使用 includePrerelease=true。

兼容修正的开发机验证目录为 `/root/workspaces/bohrium-skills-dsh-compat-20261105`，分支 `fix/dsh-compatibility-20261105`，基线 `c4f00a7b72e4780c30b93bed5e233c112864bf43`。复用原有相同版本的验证依赖，9 项 Node 检查全通过；源输入只变更兼容元数据、相关回归与文档，无需重复运行未受影响的业务或 Go 回归。

用户明确授权在 Mac 使用本机 DSH 验证。系统 CLI 已升级到 0.2.1-rc.5，Git 固定提交安装后发现并加载 18 个 Skills。另通过 `/Applications/DeepSeek Harness.app/Contents/Resources/runtime/cli/bin/dsh`（实际桌面运行时 0.2.2-rc.1），使用独立 `DSH_HOME` 安装兼容修正后的 `.tgz` 并启动宿主：Skill catalog 完整，18 个 Skills 可见，setup/job/paper 正文与 job 参考文件可读取，模型与用户调用标志均为 true。该验证没有执行模型或 Bohrium 业务请求。

本机原始证据位于 `/tmp/bohrium-dsh-local-20261105.IzmvpM/`：`result.json`（CLI）、`desktop-result.json`、`desktop-install.log`、`desktop-boot.log`。本地目录链接安装需按接入指南先安装开发依赖；正式索引使用 Git 固定 commit 安装，已验证该路径。
