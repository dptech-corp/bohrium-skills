---
name: bohrium-job
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr job 和 bohr job-group 提交、查询、停止、删除或下载计算任务。用于 BOHR_ID、机器镜像选择、科学批量计算、任务日志、排队取消、运行终止及任务组生命周期。"
---

<a id="bohrium-job"></a>

# Bohrium 计算任务

<a id="resolve-resources-and-cost-before-submission"></a>

## 提交前确定资源与费用

计算任务仅使用余额支付，按机器类型和实际用量计费。授权提交前先查看价目和可用机器：

```bash
bohr billing pricing --resource job -o json
bohr machine list --scene job --choose-type all -o json
```

确认项目、机器、镜像、命令、节点数、运行时间限制、本地输入目录、远程结果路径和费用授权后，再提交：

```bash
bohr job submit --project-id <project_id> --job_name <name> \
  --image_address <image> --machine_type <machine> --command <command> \
  --input_directory ./input --result_path /personal/results --yes -o json
```

输入为空时省略 `--input_directory`。复杂设置可以放入配置文件，但显式参数与文件中的值不得冲突。

<a id="track-jobs-by-bohr-id"></a>

## 使用 Bohr ID 跟踪任务

`bohr job list` 的 `BOHR_ID` 列是 describe、terminate、kill、delete、download 和 log 使用的公开任务标识。CLI 自动解析平台内部任务 ID：

```bash
bohr job list --running -o json
bohr job describe --id <bohr_id> -o json
```

存在归一化的 `phase` 和 `terminal` 字段时优先使用。停止请求已受理或任务仍处于非终态，均不代表操作完成。

运行中的任务使用 `bohr job terminate`；排队中的任务使用 `bohr job kill`。CLI 默认在 `--confirm-timeout` 限制内确认终态。`--no-wait` 在受理后返回，必须将结果报告为“尚未确认”。

`kill` 和 `delete` 属于破坏性操作，需要具体目标授权及 `--yes`。`terminate` 是非破坏性停止请求，仍需明确的用户意图。日志和结果会写入本地文件，下载前确认 `--out`。

<a id="job-groups"></a>

## 任务组

`bohr job submit` 可能同时返回 `jobGroupId` 和 `bohrJobGroupId`：

- 后续提交的 `--job_group_id` 使用 `bohrJobGroupId`。
- 提交后立即终止或删除任务组时，优先使用 `jobGroupId`；只有任务被索引后，`bohrJobGroupId` 才能被解析。

使用规范的 `bohr job-group` 命令创建、列出、终止、删除任务组或下载结果。任务组终止和删除均属于破坏性操作。

完整命令、参数、ID 类型、计费边界、状态与停止规则见 [命令参考](references/commands.md)。

<a id="resource-payment-boundary"></a>

## 资源支付边界

标记为 balance-only 的资源操作继续按原有规则从余额计费。全局 `--payment-source` 不会为这些资源启用光子支付；显式指定 `photon` 会在执行前被拒绝。按该资源既有的价格查询和费用授权流程操作。
