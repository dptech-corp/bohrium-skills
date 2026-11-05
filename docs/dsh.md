# DSH 接入指南

本仓库的 `skills/` 同时是独立的 DSH 原生插件包 `@dptech-corp/bohrium-dsh`，包含 17 个完整业务 Skill 和 `bohrium-setup`。它使用现有英文内容及参考文件，没有安装脚本，不运行旧的 `bohrium-skills-cli` 安装器，也不向用户 Skill 目录复制文件。

兼容范围为 **DSH 0.2.x，从 0.2.1-rc.5 起**（`>=0.2.1-rc.5 <0.3.0-0`），运行时复用宿主自带的 filesystem provider。已验证 CLI 0.2.1-rc.5 和桌面版 0.2.2-rc.1；0.3 及以后需要重新验证，原有 `bohr skill install --target dsh` 路径继续可用。

## 安装和首次使用

先通过 DSH 的 Plugins 页面，或它提供的 `dsh plugin` 命令安装。命令方式从本仓库的本地 checkout 安装：

```bash
git clone https://github.com/dptech-corp/bohrium-skills.git
cd bohrium-skills
npm install --prefix skills --ignore-scripts
dsh plugin --profile web add "$PWD/skills"
```

本地目录安装会链接 checkout，需要保留目录。示例中的 `npm install` 安装开发验证依赖，不会安装 `bohr` 或执行旧安装器；经 Git/打包产物安装时使用宿主模块。切换 checkout 或替换文件后重启 DSH。

通过 Git 安装固定版本时，从已经包含该插件的 checkout 取得完整 commit：

```bash
commit="$(git rev-parse HEAD)"
dsh plugin --profile web add "git+https://github.com/dptech-corp/bohrium-skills.git#${commit}&path:/skills"
```

该 commit 必须已在公开远端可取到。`web` 是目标 profile，可换成正在使用的 profile 名；安装会影响该 profile 的所有会话。检查 Plugins 页面里的启用状态和错误提示，按提示重启，再新建会话确认 `bohrium-setup` 和业务 Skills 可见。安装回执或文件落盘不能替代实际加载检查。

先让 Agent 运行 `bohrium-setup`。它会检查 `bohr`、引导单独安装和登录、执行 `bohr doctor`，再回到具体任务。插件不包含 CLI，也不自动给任务授权；涉及付费和写操作仍按 Skill 中的价格、预览和用户授权规则执行。

例如：让 Agent 先检查 Bohrium 环境，再读取 `bohrium-job`，解释如何查看一个已知任务的状态。真实业务请求需要可用账号、网络和相应授权。

## 通过索引发现

DSH 默认不附带索引源。需要先在 Plugins 页面开启实验性的 **Custom plugin indexes**，再进入 **Discover → Add index source** 添加索引 URL。只有添加了该源的用户，才能从这份索引搜索 Bohrium。

仓库提供手动 **Publish DSH plugin index** workflow，验证主干内容后构建 `index.json` 与 `icon.svg`，再通过 GitHub Pages 发布。首次使用需在仓库 Pages 设置中选 GitHub Actions 作为发布来源；确认仓库没有其他 Pages 站点需要共用后再运行。workflow 的 deployment URL 加 `index.json` 即实际索引地址。默认项目站点通常是：

```text
https://dptech-corp.github.io/bohrium-skills/index.json
```

以上地址是部署后的预期地址，不表示当前已上线。自定义域名以实际 deployment URL 为准；索引地址需要直接返回 2xx，不经过重定向。发布一次索引不会让所有 DSH 用户默认发现，也不代表进入官方内置目录。

维护者可先生成本地静态产物用于检查：

```bash
node scripts/build_dsh_index.mjs \
  --commit "$(git rev-parse HEAD)" \
  --timestamp "$(git show -s --format=%ct HEAD)" \
  --output _site/index.json
```

索引固定 Git commit 和 `skills` 子目录，不使用移动分支。包版本与根 `package.json` 保持一致；发布新版本时同步两者。索引中的 DSH 兼容范围来自插件包，需在对应运行时验证后再扩展。正式索引只能指向已经包含全部目标文件的公开 commit。

## 更新、禁用和卸载

在 Plugins 页面关闭插件可停止贡献这些 Skills。要更新 Git 来源，选择一个已验证并公开的新 commit，使用上面的固定 commit 安装命令替换来源；Git 来源的更新不依赖 npm 自动更新检测。替换插件代码后重启 DSH，并在新会话核对版本与 Skills。

命令行卸载：

```bash
dsh plugin --profile web remove @dptech-corp/bohrium-dsh
```

移除仅取消该插件的注册和包依赖，个人及项目中的 Skills 继续保留。按 DSH 提示重启后核对；用户以前通过直接安装方式留下的同名 Skill 也会保留。要确认移除结果，请区分插件贡献和这些已有副本。

## 验证范围

CI 针对实际 `npm pack` 产物检查完整文件，使用真实 Cordis/DSH registry 验证发现、按名加载、资源路径、用户覆盖以及禁用和重新启用，使用官方索引 validator 校验生成结果。测试不调用业务 API，也不证明浏览器图标已正确渲染、公共 Pages 已上线或真实付费任务已经完成；这些结果需分别记录。

## English

The independent `@dptech-corp/bohrium-dsh` package lives in `skills/` and contains all 17 workflow Skills plus `bohrium-setup`, including their reference files. It has no install scripts, does not invoke the legacy installer, and does not copy files into user Skill directories. Supports **DSH 0.2.x from 0.2.1-rc.5** (`>=0.2.1-rc.5 <0.3.0-0`), including Desktop 0.2.2-rc.1; 0.3 and later require validation.

Clone this repository, enter its root, and install from an absolute local path:

```bash
npm install --prefix skills --ignore-scripts
dsh plugin --profile web add "$PWD/skills"
```

Keep the checkout when using a linked local installation. For a public commit that contains the bundle, use an immutable Git source:

```bash
commit="$(git rev-parse HEAD)"
dsh plugin --profile web add "git+https://github.com/dptech-corp/bohrium-skills.git#${commit}&path:/skills"
```

Replace `web` with the intended profile. Check Plugins for activation errors, restart when requested, and open a new session. Run `bohrium-setup` first to install and sign in to `bohr` separately. The plugin does not authorize spending or changes.

For discovery, enable experimental **Custom plugin indexes**, then use **Discover → Add index source**. The manual **Publish DSH plugin index** workflow validates main and deploys a static index through GitHub Pages. Configure Pages to use GitHub Actions first, and ensure it will not replace an existing site. Use the actual deployment URL followed by `index.json`. The expected default URL is `https://dptech-corp.github.io/bohrium-skills/index.json`; it is available only after deployment. DSH does not ship this source by default.

Disable the bundle through Plugins. To update a Git installation, install a verified public commit using the command above and restart DSH; Git sources do not use npm update detection. Remove it with:

```bash
dsh plugin --profile web remove @dptech-corp/bohrium-dsh
```

Personal/project Skills remain intact, including same-name Skills previously installed directly. CI validates the packed content, real registry lifecycle, and official index format. Public hosting, rendered UI, and real business requests are separate acceptance checks.
