---
name: bohrium-project
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr project 查询、创建或删除项目，并查询成员；为数据集、文件、计算任务、节点、镜像、Notebook 或 Sandbox 确定项目 ID 与创建额度。"
---

<a id="bohrium-project"></a>

# Bohrium 项目

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 选择计算或存储归属 | 列表 → 查看项目及成员 → 将准确项目 ID 交给下游 |
| 按限额创建项目 | 确认名称和限额 → 创建 → 用 list 和 get 定位新项目 |
| 删除项目 | 检查成员及依赖 → 准确目标授权 → 删除并核实 |

## 工作流：创建资源前先确定项目

1. 列出项目，以 `data.items[].id` 选择，再查看该 ID 的详情和成员：

```bash
bohr project list -o json
bohr project get <project_id> -o json
bohr project members <project_id> -o json
```

2. 核对返回的身份和归属，不能仅按可能重名的显示名称选择。成员 user ID 不是 project ID。先创建时查看实际响应；它是上游数据，不能保证固定 `data.projectId`。按预期的唯一名称重新列出，并用 `get` 核实后再继续。
3. 将同一项目 ID 交给各资源操作：

| 消费者 | 项目输入 |
|---|---|
| Dataset create | `--pid <project_id>` |
| Job submit／Node create／Image build | `--project-id <project_id>` |
| 项目共享盘 File 操作 | `share/... --project-id <project_id>` |

4. 完成标准是已确认项目与资源的绑定，项目列表成功还不够。继续遵守下游 Skill 的费用和写入规则。详情失败时，先区分权限、网络或上游结果不完整与项目缺失，再选择或创建替代项目。

在 Dataset、File、Job、Node、Image、Notebook 或 Sandbox 流程使用项目之前，先确定项目及其 ID：

```bash
bohr project list -o json
bohr project get <project_id> -o json
bohr project members <project_id> -o json
```

项目命令当前免费。create 是写操作，支持在创建时指定费用上限：

```bash
bohr project create --name research-project \
  --total_cost_limit 1000 --month_cost_limit 200 -o json
```

`0` 的含义以平台响应为准，不能自行解释为无限制或禁用。保留创建返回的项目 ID；项目名称、成员用户 ID 和项目 ID 不能互换。

delete 是破坏性操作，可能影响项目资源的访问。先检查项目与成员，识别相关共享路径或计算资源，再获取对具体目标的授权：

```bash
bohr project delete <project_id> --yes -o json
```

当前 CLI 不能重命名项目、创建后修改额度，也不能添加、移除、提升、降级或恢复成员。不要将上游服务操作转化为虚构的 CLI 参数。

完整命令、创建输入、项目 ID 交接和删除边界见 [命令参考](references/commands.md)。
