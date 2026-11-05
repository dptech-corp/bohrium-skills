# Bohrium Skills 的 DSH 插件接入

## Goal

DSH 用户能将现有完整 Bohrium Skills 作为一个原生插件安装、启用和移除，并通过自行添加的索引源发现该插件。

## Current evidence

基线为 origin/main 688672f42acdc6c3582286610e6944318ff38acc。`skills/` 已有 17 个完整业务 Skill 和 `bohrium-setup`；根 npm 包的 `scripts/postinstall.js` 会下载安装器并同步用户目录，不适合直接作为 DSH bundle。`scripts/check_dsh_consumer.mjs` 只检查文件系统 provider，不覆盖插件生命周期。DSH 官方 host-plugin 和 plugin-market 文档规定 `dsh.bundle.patch`、导出显示资源、Git 固定 commit 和子目录索引。已检查的 filesystem provider 0.1.5-rc.2 和目标运行时 0.2.1-rc.5 支持 `includeDefaultRoots:false` 和 `bundledSkillDir`。

## Behavior scenarios

用户安装仓库 `skills` 子目录提供的独立包后，启用插件即可加载 18 个英文 Skills，参考资源仍指向包内原文件。运行 setup skill 引导单独安装并登录 bohr。禁用或卸载仅移除插件贡献，保留个人及项目 Skills。用户添加 HTTPS 索引后可发现条目；未添加源时不承诺默认可发现。

## Requirements

- 使用 `skills/` 本身作为独立 DSH 包，无安装脚本，不复制或修改 Skill 正文；复用 DSH 的 filesystem provider 解析和加载内容。
- provider 只贡献包内 bundled Skills，保留项目及用户同名 Skill 的优先级，并在卸载时释放注册与监听资源；原生包兼容 DSH 0.2.x，从 0.2.1-rc.5 起（`>=0.2.1-rc.5 <0.3.0-0`），包含桌面版 0.2.2-rc.1；更旧版本与 0.3 以后的系列需单独验证。
- 独立包名称为 `@dptech-corp/bohrium-dsh`，版本沿用仓库版本；提供英文、中文显示元数据及原创 SVG 图标。
- 索引由构建脚本产生，使用 Git HTTPS 地址、完整 commit、`path: skills`、确切包版本和 Unix 秒时间戳。拒绝缺失或格式错误的 commit、时间戳和版本漂移。
- CI 验证真正打包后的资源完整性、真实 registry 的加载和禁用/卸载语义、官方索引校验；手动 Pages workflow 负责 HTTPS 托管，不在开发过程中发布或修改线上设置。

## Decisions

使用独立子目录包以隔离旧安装器副作用。直接复用既有 filesystem provider，不增加自制 frontmatter 解析器。保持原有三客户端安装器和 Agent Plugin、Gemini 入口。实现范围已经用户确认，spec 不引入额外产品决策。工程风险按 Deep 记录，涉及外部插件契约与一个宿主自带 DSH 模块的使用；验证在任务独占开发机目录进行。

## Acceptance criteria

- packed npm 产物包含 18 个 Skill 的完整文件树、入口、patch、locale、图标，没有旧安装器和生命周期脚本。
- 真实 Cordis/DSH registry 安装插件后能按名加载完整正文及资源；释放插件后贡献消失；其他用户 Skill 不受影响。
- 生成的索引经官方 validator 接受，源字段精确指向选定 commit 的 skills 子目录；错误输入失败。
- 中文、英文 README 提供直接安装、索引开启/添加、setup、固定版本更新和卸载说明，准确标注预览兼容范围。
- 本次没有对公开索引、npm 或用户正在使用的 DSH profile 做写操作。UI 渲染、真实业务请求和公开托管的结果与代码验证分开报告。

## Scope and non-goals

本仓增加 DSH 原生适配、索引构建、手动托管工作流和验证。CLI、MCP、原有业务 Skill、飞书原文及其他仓库不在修改范围。不自动安装 bohr，不执行付费请求。

## Verification

在 root@39.100.68.53 的 /root/workspaces/bohrium-skills-dsh-20261105 验证，先检查 HEAD、分支、dirty、同步和工具版本。使用 node:test 在已同意的插件加载/生命周期、索引和发布物边界验收；先失败再实现。真实 DSH CLI 在隔离 DSH_HOME/profile 中验证安装、加载和移除，无法证明的环节明确留证。执行既有 Go/Node/Python 检查中受影响的检查并检查最终 diff。
