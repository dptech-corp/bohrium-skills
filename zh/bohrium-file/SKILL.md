---
name: bohrium-file
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr file 浏览和管理 personal/share 文件盘，上传下载、复制移动、创建目录、查看状态或删除文件，处理分片续传及项目路径。"
---

<a id="bohrium-file"></a>

# Bohrium 文件盘

使用规范路径前缀：

- `personal/<path>` 表示当前用户的个人文件盘，要求 `--project-id 0`，这也是默认值。
- `share/<path>` 表示项目共享文件盘，要求提供所属项目的正整数 `--project-id`。

修改前先确定目标：

```bash
bohr file list personal/ --limit 50 -o json
bohr file stat personal/results/run.log -o json
bohr file list share/project-data --project-id <project_id> -o json
```

规范的写入和传输命令如下：

```bash
bohr file upload ./input.dat personal/jobs/input.dat -o json
bohr file upload ./large.tar personal/jobs/large.tar --upload-concurrency 8 -o json
bohr file upload ./large.tar personal/jobs/large.tar --upload-concurrency 8 --resume -o json
bohr file download personal/jobs/output.dat --dest ./output.dat
bohr file download personal/jobs/results/ --recursive --dest ./results
bohr file mkdir share/new-run --project-id <project_id> -o json
bohr file copy personal/a personal/b -o json
bohr file move share/old share/new --project-id <project_id> -o json
```

上传或下载前确认具体本地源文件、目标文件和远程路径。copy 和 move 默认只影响一个对象，只有显式指定 `--recursive` 时才作用于整个前缀。

大于等于 50 MiB 的上传使用 25 MiB 分片。通过 `--upload-concurrency` 限定全局请求数，范围 1–16，默认 5。中断后保持本地文件和远程路径不变，添加 `--resume`，CLI 会复用已完成分片。进度写到 stderr，机器可读的 stdout 仍是一个结果信封。File append 契约要求单个对象的分片有序，所以即使配置上限更高，其实际并发峰值仍为 1。较小的二进制文件会从字节零重新上传。

删除属于破坏性操作。先 list/stat 精确目标，只有明确选中了整个前缀时才递归删除：

```bash
bohr file delete share/obsolete --project-id <project_id> --recursive --yes -o json
```

当前所有 File 命令免费，但写入、本地覆盖、移动和删除仍需用户意图。文件盘空间只由 `personal/` 或 `share/` 路径前缀决定。

当前 CLI 不提供解压、服务端文件名搜索或内容搜索。完整命令、参数、上传结果字段、恢复警告、路径约束和确认边界见 [命令参考](references/commands.md)。
