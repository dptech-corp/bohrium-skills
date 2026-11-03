---
name: bohrium-node
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr node 创建、查询、连接、重启、停止或删除交互式开发节点。用于开发机、长期 CPU/GPU 资源、机器价格、镜像、磁盘、自动关机、SSH 和节点生命周期；非交互批量计算使用 Job。"
---

<a id="bohrium-node"></a>

# Bohrium 开发节点

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 打开交互式开发环境 | 资源和价格 → 创建 → 等待运行 → 连接 |
| 复用已有环境 | 确定 Node ID → 检查状态 → 停机时按授权重启 |
| 结束使用 | 保存工作 → 停机 → 确认已停止；明确废弃时才删除 |

## 工作流：创建、使用并停止开发机

1. 按下方费用预检选择项目、镜像、机型、磁盘和自动关机时间。获授权创建后，将 `data.id` 保存为 Node ID；从列表选择时使用 `data.items[].nodeId`。
2. 在约定的等待预算内，用同一 Node ID 查询至 `data.status=2` 再连接；到期仍未就绪时保留该 ID，并报告就绪尚未确认：

```bash
bohr node get <node_id> -o json
bohr node connect <node_id>
```

3. `get` 在 `data.requestedNodeId` 保留查询 ID，详情中的 `nodeId` 或 `machineId` 可能属于底层机器。生命周期及 SSH 操作继续使用原 Node ID。即使选择全局 JSON 输出，`connect` 仍启动交互式 SSH；应使用交互终端并保护连接凭据。
4. 保存用户工作，按请求停止：

```bash
bohr node stop <node_id> -o json
bohr node get <node_id> -o json
```

5. 完成要求目标环境实际可用；获授权停机后，还要从详情或列表观察到 `status=-1`，仅受理停机不算完成。启动或 SSH 失败时先检查原开发机，再决定是否创建另一台计费资源。重启可能恢复收费；删除是独立的破坏性决策。

Node 是交互式、可能长期运行的计算资源。创建或重启前先查看机器、价格和资源：

```bash
bohr billing pricing --resource node -o json
bohr machine list --scene node --choose-type all -o json
bohr node resources -o json
```

创建和重启仅使用余额支付。添加 `--yes` 前确认项目、机器、镜像、磁盘、自动关机和费用：

```bash
bohr node create --name dev-node --project-id <project_id> \
  --image_path <image> --machine_type <machine> --disk_size 40 \
  --turnoff_after 24 --yes -o json
```

磁盘大小必填，只能是 **20、40、80、100 GB**。`--turnoff_after` 默认 `24` 小时，常用值为 `1`、`8`、`24`，`-1` 表示永不自动关机。无限期运行的资源需要尤其明确的费用意图。也可通过 JSON 配置文件输入，提交前先检查内容。

始终保留创建返回的 ID 作为本次请求的 Node ID。后续详情可能另有内部 machine/node ID，不能静默替换生命周期命令中的创建 ID。

按 Node ID 跟踪和连接：

```bash
bohr node list --started -o json
bohr node get <node_id> -o json
bohr node connect <node_id>
```

connect 会在本地启动交互式 SSH 进程。凭据和连接细节不应进入日志或总结。

生命周期操作：

```bash
bohr node stop <node_id> -o json
bohr node restart <node_id> --turnoff_after 24 --yes -o json
bohr node delete <node_id> --yes -o json
```

stop 是写操作；restart 可能恢复计费；delete 是破坏性操作。核对返回状态，不能把请求受理当成生命周期完成。删除后确认唯一节点名称或 ID 已从 `bohr node list` 消失。`bohr node get` 在删除后可能短暂返回旧详情，仅凭旧记录不能证明节点仍在运行。

当前 CLI 没有独立的 start、rename 或 dataset-binding 子命令。

完整命令、过滤条件、资源输入、计费边界和生命周期约束见 [命令参考](references/commands.md)。

<a id="resource-payment-boundary"></a>

## 资源支付边界

标记为 balance-only 的资源操作继续按原有规则从余额计费。全局 `--payment-source` 不会为这些资源启用光子支付；显式指定 `photon` 会在执行前被拒绝。按该资源既有的价格查询和费用授权流程操作。
