---
name: bohrium-image
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr image 发现、查看、拉取、推送、构建或删除容器镜像。用于本地 Docker 上下文、远程 Dockerfile 构建、沙箱快照、构建日志、镜像状态与生命周期。"
---

<a id="bohrium-image"></a>

# Bohrium 容器镜像

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 复用已有镜像 | 列出或搜索 → get → 保存规范仓库地址 |
| 使用本地源码构建 | 本地构建 → 返回的 `pushCommand` → 核实推送回执 |
| 远程构建 Dockerfile 文本 | 远程构建 → 镜像终态 → 检查最终镜像 |
| 保存沙盒根文件系统 | 下方存储门禁 → commit → 终态；挂载数据单独处理 |

## 工作流：本地构建、发布，再交给计算资源

1. 确认 Docker 上下文和 daemon、Dockerfile 与构建目录、唯一短名称、项目及写入意图。本地模式会预留远端记录和配额，但不会推送。

```bash
bohr image build --mode local --dockerfile ./Dockerfile --context .   --name research-env:v1 --project-id <project_id> -o json
```

2. 保留本地构建结果的 `data.localImage`、`imageUrl`、`id` 和 `pushCommand`。`uploaded=false` 表示只完成本地构建。使用同一 Docker daemon、Bohrium 身份或 profile 及项目，执行实际返回的 `pushCommand`，不能自行拼接仓库路径。
3. 成功推送会校验最终 `data.id`、`imageUrl`、`status=1`、`digest` 和 `sizeBytes`。将最终 `imageUrl` 交给 Job 的 `--image_address` 或 Node 的 `--image_path`；镜像记录 ID 和本地标签都不是仓库地址。
4. 本地 prepare 记录不是远端构建任务，不能用 `status` 或 `build-log` 轮询。远程构建或 commit 使用下方独立的 `0/1 → 2/3/4` 生命周期；其成功状态 `2` 不能套用于本地推送回执。
5. 交付已发布地址、digest 及选定消费者需要的条件。构建或推送中断、prepare 结果未知时，先按参考核对同一名称和记录，再预留新记录。只有本地构建镜像还未完成发布。

有本地上下文的 Docker 构建使用 `build --mode local`，之后单独执行返回的推送命令。默认 `build --mode remote` 只把 Dockerfile 文本发给远程构建器。`commit` 在服务端保存运行中沙箱容器的根文件系统快照，不能假设挂载卷会包含在内。

创建镜像前先发现已有镜像：

```bash
bohr image list --limit 50 -o json
bohr image search "DeePMD" -o json
bohr image get <image_id> -o json
```

`--limit` 范围为 `1-512`，默认 `512`；只需有限目录样本时使用更小的值。

pull 将 Bohrium 镜像复制到本地 Docker daemon。先确认 Docker 已安装并运行，且镜像地址属于 Bohrium registry：

```bash
bohr image pull <bohrium_image_address> --timeout 30m -o json
```

使用本地上下文构建，再在相同 Docker context/daemon、Bohrium profile 和身份、项目及短名称下执行返回的 `pushCommand`：

```bash
bohr image build --mode local --dockerfile ./Dockerfile --context . \
  --name my-image:v1 --project-id 1234 -o json
```

该流程会通过 prepare 预留远程镜像记录和配额，并非离线操作。本地构建不提交远程计算、不要求 `--yes`，也不会自动推送。本地成功后执行 `pushCommand`；不要用 prepare 记录 ID 轮询 `image status` 或 `build-log`。使用唯一名称，不要并发执行同名操作。明确的构建失败会尝试 abort；prepare 未返回可靠结果时，先检查状态，错误也可能留下记录，不能假设记录不存在或已被清理。

也可使用 AccessKey 推送 Docker daemon 中已有镜像。此命令调用 `PATH` 上的 Docker CLI，不会构建镜像，也不会上传 Dockerfile 或上下文。先自行使用 `docker build`、`docker load` 或 `docker pull` 准备镜像。第二个参数是与 build/commit 相同风格的短镜像名；项目 ID 来自 `--project-id` 或 `BOHR_PROJECT_ID`：

```bash
docker build -t my-image:v1 .
bohr image push my-image:v1 my-image:v1 --project-id 1234 --timeout 30m -o json
```

registry 目标只使用 prepare 响应中的规范 `imageUrl`，不能自行拼接项目/用户路径。生命周期和失败恢复见 [命令参考](references/commands.md)。

<a id="rootfs-gate-before-a-sandbox-snapshot-commit"></a>

## 沙箱快照 commit 前的 rootfs 检查

每次 `bohr image commit` 提交前，立即执行 `bohr sandbox storage inspect <sandbox_id> -o json`。不能等 commit 开始后再依靠 80% 阈值扩容：commit apply 消耗临时 rootfs 的速度可能快于扩容进入 Ready。按 [命令参考](references/commands.md) 的失败数据公式估算容量，必要时先扩容，只有 `status=ready` 后才继续。预期未压缩快照增量未知时，不能当作零或声称容量安全。rootfs 只能扩容，不能缩容。

这是 Agent 的前置流程，并非 CLI 自动执行的行为，也不适用于 `bohr image build --dockerfile`。

远程 build 和 commit 按 Sandbox 价目从余额收取资源费用。先查看价格结构并获得费用授权：

```bash
bohr billing pricing --resource sandbox -o json
bohr image build --mode remote --dockerfile ./Dockerfile --name project/image:v1 \
  --project-id <project_id> --wait --yes -o json
bohr image commit --sandbox-id <sandbox_id> --name project/snapshot:v1 \
  --project-id <project_id> --wait --yes -o json
```

使用不可变名称，不要使用可变的 `latest` 标签。远程 build 只发送所选 Dockerfile 文本，不发送本地构建上下文；commit 保存所选运行中沙箱的容器根文件系统快照。两者都不使用本地 Docker 镜像。确认具体本地文件、沙箱、所属项目、名称和费用授权。

没有 `--wait` 的远程构建或快照提交，应保留返回的镜像 ID，再查询：

```bash
bohr image build-log <image_id> --follow
bohr image status <image_id> -o json
bohr image dockerfile <image_id>
```

Creating（`0`）和 pending（`1`）是非终态；`2` 表示成功，`3` 表示失败，`4` 表示取消。`--wait` 在 `2/3/4` 停止，用 `--wait-timeout` 限定等待。本地超时不能证明远程构建已停止。

delete 是破坏性操作：

```bash
bohr image delete <image_id> --yes -o json
```

当前 CLI 没有 share/unshare 或元数据更新命令。完整命令、参数、本地/远程影响、计费和生命周期完成条件见 [命令参考](references/commands.md)。

<a id="resource-payment-boundary"></a>

## 资源支付边界

标记为 balance-only 的资源操作继续按原有规则从余额计费。全局 `--payment-source` 不会为这些资源启用光子支付；显式指定 `photon` 会在执行前被拒绝。按该资源既有的价格查询和费用授权流程操作。
