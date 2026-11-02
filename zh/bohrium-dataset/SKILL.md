---
name: bohrium-dataset
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr dataset 管理数据集，发现数据集和版本标识，创建并上传本地目录、断点续传、下载或删除数据集。"
---

<a id="bohrium-dataset"></a>

# Bohrium 数据集

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 发布本地数据 | 确认归属和来源 → 创建上传 → 检查数据集及版本 |
| 获取可复现输入 | 确定 dataset ID → 选择就绪版本 → 下载并校验 ZIP |
| 恢复中断传输 | 同一上传身份加 `--resume`，或保留下载版本和目的路径 |

## 工作流：发布并取回准确的数据集版本

1. 确认项目、本地目录树和目标数据集身份，再使用下方创建命令。保存 `data.dataset_id`，若返回 `data.version_id` 也一并保留；数据集 ID 和版本 ID 各有用途。
2. 检查数据集及已有版本：

```bash
bohr dataset get <dataset_id> -o json
bohr dataset versions <dataset_id> -o json
bohr dataset download <dataset_id> --version-id <version_id> --dest ./dataset.zip -o json
```

3. 从 `versions` 中所选记录的 `data.items[].id` 填入 `<version_id>`。可下载版本应满足 `status=committed`、`packStatus=completed`，且有可用 `downloadUri`。上传已受理或仍在打包的版本尚不是就绪归档；在等待预算内继续检查该版本。
4. 下载准确版本，仅在回执 `data.verified=true` 时认定完成；报告 `path`、`bytes` 及 dataset/version 绑定。下游明确需要解压前，将 ZIP 保留为归档。
5. 上传中断时保留四项身份输入，按下方规则加 `--resume`。下载中断时保留版本和目标路径，使 `.part` 恢复绑定同一数据。`--force` 会丢弃恢复状态，需要覆盖意图；更换版本不能算原部分下载的安全继续。

传输或删除数据前，先通过读取命令确定稳定 ID：

```bash
bohr dataset list --number 50 --title "training" -o json
bohr dataset get <dataset_id> -o json
bohr dataset versions <dataset_id> -o json
```

创建数据集会把选定的本地路径上传到 Bohrium。写入前确认具体路径、数据集身份及所属项目：

```bash
bohr dataset create --name training-set --path training-set \
  --pid <project_id> --lp ./data --upload-concurrency 5 -o json
```

中断上传只能在 `--name`、`--path`、`--pid` 和 `--lp` 保持不变时恢复。添加 `--resume` 复用兼容的缓存状态。`--upload-concurrency` 的全局范围为 `1-16`；增大它会增加并发 HTTP 请求，不会增加数据集版本数。

下载自动使用稳定的 `.part` 文件支持断点续传：

```bash
bohr dataset download <dataset_id> --version-id <version_id> --dest ./dataset.zip
```

只有用户希望丢弃现有目标文件和部分下载产物、从头开始时，才使用 `--force`。

数据集操作当前免费，但 create 会上传用户数据，delete 属于破坏性操作。执行 `bohr dataset delete <dataset_id> --yes` 前核对具体 ID 和删除意图。免费不等于获得修改授权。

当前 CLI 能列出已有版本，不能通过 dataset 子命令创建新版本、编辑元数据、管理权限或查询配额。

完整命令、参数、续传不变量、本地文件影响和删除边界见 [命令参考](references/commands.md)。
