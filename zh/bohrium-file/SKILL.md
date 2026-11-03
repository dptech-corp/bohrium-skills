---
name: bohrium-file
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr file 浏览和管理 personal/share 文件盘，上传下载、复制移动、创建目录、查看状态或删除文件，处理分片续传及项目路径。"
---

<a id="bohrium-file"></a>

# Bohrium 文件盘

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 收发一个云盘对象 | 确定空间、项目和路径 → 传输 → 检查 stat 或回执 |
| 复制或下载目录树 | 列出最小范围前缀 → 显式递归操作 → 检查逐文件结果 |
| 恢复上传 | 保留本地文件和远端目标 → 检查未知结果 → 恢复兼容会话 |

## 工作流：传输项目输入并核实往返结果

1. 确认本地源和目标、项目归属及远端键。share 路径的每一步都需要同一个正数项目 ID。

```bash
bohr file upload ./input.dat share/experiment/input.dat --project-id <project_id> -o json
bohr file stat share/experiment/input.dat --project-id <project_id> -o json
bohr file download share/experiment/input.dat --project-id <project_id> --dest ./retrieved-input.dat -o json
```

2. 上传成功返回 `data.remote_path`、`space`、`project_id` 和 `size`。将完整 `remote_path` 和项目范围交给 stat 或 download；只保留文件名会丢失存储路由。核对远端元数据是否对应所需对象及预期大小。
3. 下载成功报告 `data.dest_path` 和 `size`。交付已核实的远端目标及实际本地路径；关键内容在有条件时进一步比对字节或校验和，仅大小一致不能证明内容完整。
4. 递归下载检查 `downloaded`、`skipped` 和 `failed`，不能把 `files` 当成本次新传输数量。存在失败路径时报告部分交付，目录非空不等于完成。
5. `UPLOAD_NOT_CONFIRMED` 时，先检查同一远端对象及返回的恢复提示，再做新写入。兼容的上传中断按下方分片规则恢复。下载失败会保留原有完整目标；重试属于读取及获授权的本地文件写入。

使用规范路径前缀：

- `personal/<path>` 表示当前用户的个人文件盘，要求 `--project-id 0`，这也是默认值。
- `share/<path>` 表示项目共享文件盘，要求提供所属项目的正整数 `--project-id`。

修改前先确定目标：

```bash
bohr file list personal/ --limit 50 -o json
bohr file stat personal/results/run.log -o json
bohr file list share/project-data --project-id <project_id> -o json
```

`list` 每次读取一页。需要完整目录时，保持路径、项目和 limit 相同；`has_more=true` 时将 `data.pagination.next_token` 交给 `--next-token`：

```bash
bohr file list personal/ --limit 50 --next-token "<next_token>" -o json
```

`has_more=false` 或达到用户所需数量时停止。`has_more=true` 但游标缺失或重复时，应报告覆盖不完整，不能重复读取首屏。

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
