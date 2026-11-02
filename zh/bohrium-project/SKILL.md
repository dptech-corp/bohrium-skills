---
name: bohrium-project
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr project 查询、创建或删除项目，并查询成员；为数据集、文件、计算任务、节点、镜像、Notebook 或 Sandbox 确定项目 ID 与创建额度。"
---

<a id="bohrium-project"></a>

# Bohrium 项目

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
