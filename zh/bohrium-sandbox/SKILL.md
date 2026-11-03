---
name: bohrium-sandbox
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr sandbox 管理 Bohrium 云沙箱的生命周期、配额、命令执行、PTY、文件、费用、模板、镜像构建/快照、文渊数据挂载与软件安装。用于云沙箱操作、连接超时、配额或模板容量不足等问题。"
---

<a id="bohrium-sandbox-bohr-sandbox"></a>

# Bohrium 云沙箱（`bohr sandbox`）

## 按任务选择路径

| 目标 | 所需参考或路径 |
|---|---|
| 运行一次计算并取回文件 | Doctor 和配额 → 创建 → 上传 → 前台执行 → 下载 → 按授权清理 |
| 长时间或后台计算 | [执行模式](references/sandbox/execution-modes.md)及[执行恢复](references/sandbox/exec.md) |
| 读写 Wenyon 数据集 | [实例生命周期和结算](references/sandbox/lifecycle.md#mount-a-wenyon-dataset-in-an-instance) |
| 安装软件、扩容或保存镜像 | 执行前读取对应软件、存储或镜像参考 |

## 工作流：运行脚本并交付输出

1. 检查 `doctor` 的 `data.ready`、配额、模板与资源及已授权费用。确认脚本、远端路径和本地输出位置。本例使用普通、无 Wenyon 挂载的沙盒，创建请求 ID 只生成一次，并在工具调用间保留：

```bash
REQUEST_ID="$(uuidgen)"
bohr sandbox create sdbxagent --timeout 3600 --request-id "$REQUEST_ID" --yes -o json
```

2. 用成功响应规范字段 `data.sandbox_id` 填入 `<sandbox_id>`，`sandboxID` 是兼容拼写。回执或错误包装按下方流程核对；operation ID 不是 sandbox ID。

```bash
bohr sandbox files write --source ./run.py <sandbox_id> /tmp/run.py -o json
bohr sandbox exec <sandbox_id> -o json -- python /tmp/run.py
bohr sandbox files read <sandbox_id> /tmp/result.json --destination ./result.json -o json
```

3. 脚本必须实际生成 `/tmp/result.json`。前台执行要求 `data.result_known=true` 且 `data.exit_code=0`，再检查所需输出。后台回执 `start_confirmed=true`、`result_known=false` 只确认启动。argv 模式下，全局 CLI 参数放在 `--` 前，之后的所有内容都是远端参数。
4. 清理前，将下载返回的 `output`、传输统计和 manifest 与所需本地产物核对。输出已安全取回且清理获授权后再删除；Wenyon 写挂载还需要下方明确的 commit 或 discard 结算。
5. 创建结果未知时，使用原请求 ID 和身份范围核对；执行结果未知时，按执行参考使用返回的进程或请求证据。重建沙盒或重跑有副作用的脚本可能重复工作。本地等待截止不能证明远端任务已结束。

通过 `bohr sandbox` 管理 Bohrium 平台按需创建的云沙箱，覆盖生命周期（`create`、`exec`、`delete`、`list`、`describe`）、配额自查、交互式 PTY、rootfs 检查和扩容、文件传输、模板、容器镜像 commit/build，以及沙箱内安装软件。

这不是 E2B SDK，不需要 `E2B_API_KEY`，也不调用 `api.e2b.dev`。操作在 Bohrium 平台上执行，使用 Bohrium AccessKey 鉴权。创建带 Wenyon 实例挂载的沙箱，还需要由指定环境变量提供 Wenyon/Vouch 用户 JWT，AccessKey 不能替代。

> **代理开关。** 沙箱默认关闭出站 HTTP 代理。访问 PyPI `pypi.org`、GitHub、HuggingFace 等海外来源时，可按 [网络参考](references/sandbox/network.md) 配置当前环境批准的代理，完成后关闭。公开文档中的 `proxy.example.com` 是占位符，不能原样执行。镜像内的国内 Aliyun 源不受影响。

> **默认行为陷阱：ML 或长期任务先读。** 根盘基础容量是 **30Gi**，额外容量在创建模板时申请，SKU 没有磁盘字段；每用户 running-sandbox quota 统计 Running+Hibernating，创建前先用 `bohr sandbox quota -o json`，达到上限时先释放名额。`exec` 没有执行截止时间，但网络故障和沙箱有效期仍会影响任务；二进制下载使用 `files read --destination` 保留原始字节，stdout 字节模式输出 base64；`delete` 不可逆，已停止沙箱的文件不能读取；代理默认关闭；模板拒绝可变的 `:latest` 标签。操作前阅读 [陷阱与训练检查表](references/sandbox/pitfalls.md)。

<a id="agent-rules"></a>

## Agent 操作规则

- create 结果不确定或响应丢失时，不要盲目重建。保持原 profile 和身份，使用 `bohr sandbox describe --create-request-id` 按原 request ID 核对。单次 404 不能证明没有创建成功。模板准备失败先检查模板状态；只有明确存在实例时才检查失败实例记录。
- Wenyon 的 `--wenyon-dataset-id`、`--wenyon-mount-path`、`--wenyon-mode`、`--wenyon-version` 传给 **sandbox create**，使用普通计算模板。默认 read/current 固定已有版本；write/current 预载已有内容；仅 write/none 从空目录开始。`--wenyon-user-token-env WENYON_USER_JWT` 是环境变量名，不是 token 字面量；`BOHR_ACCESS_KEY` 不能替代。不能与 `--inherit-auth` 组合。可写挂载以 `delete --commit` 或 `delete --discard` 结束；保存时使用 `--wait` 并核对 `produced_version`。create 不确定时先按 request ID 恢复。恢复截止时间和凭据规则见生命周期参考。
- `sandbox create --inherit-auth` 默认关闭，只有用户明确授权将调用方登录态传入沙箱时才能使用。root 和获准读取私有 tmpfs 包的进程都能使用继承凭据；AccessKey 与 Vouch 均要求经过静态校验的 npm bohr-cli 2.6.45+ 布局。Vouch 不包含 refresh token，随 access token 过期。注入返回 `PARTIAL_FAILURE` 时删除返回的沙箱，不检查或复用，也不重新 create；子类型 `sandbox_auth_rollback_uncertain` 要立即删除。`sandbox_create_outcome_uncertain` 则先按 request ID 列出并核对。
- `template create --shm-size` 报服务端未回显 `shm_size` 时，模板可能已存在。先检查模板，后端升级或修正后再按明确意图删除或重建，不要盲重试。
- `/share`、`/personal` 可能是网络文件系统，避免反复递归 `find`、`du`、`grep -R` 或内容搜索。反复按文件名查找时，对最窄范围建立私有、仅含路径的本地索引，写入/挂载变化或 10 分钟后刷新，修改前实时核对路径。见 [远程存储索引](references/sandbox/remote-storage-index.md)。
- 逐沙箱费用通过 `bohr sandbox list --query all` 查看，`COST` 每行有自己的单位。`paymentType=1` 表示光子，`0` 或缺失表示 CNY；光子行优先 `photonCost`，再回退 `cost`；现金行使用 `cost`。`cost` 缺失且没有可用 `photonCost` 表示不可得，不能编造零；显式零才表示查询时已知为零。JSON 保留字段及缺失情况。这些是任务成本，不是 `notice.billing.balance`，结算可能滞后于生命周期变化。
- 命令和新 PTY 默认以 `root` 运行；只有用户要求时才使用 `--user`，已存在 PTY 保持原用户。
- 位置参数形式的 `exec` 保留 argv 边界；管道、展开和重定向通过 `-c/--command` 交给 shell，用 `--` 结束本地参数解析。未知结果恢复和保留脚本规则见 [exec 参考](references/sandbox/exec.md)。
- 每次非 Dockerfile 快照 `bohr image commit` 提交前，立即执行 `bohr sandbox storage inspect <sandbox_id> -o json`。命令组被禁用时按存储参考在沙箱内对真实写路径运行 `df -Pk`，这只能测容量，不能启用扩容。commit apply 可能短时间消耗 rootfs，不能等使用率达到 80% 再扩容。依据实时 `used_bytes` 和可证明的未压缩 commit 增量估算总容量，未知增量不能当作零。需要扩容时必须有 `capabilities.dynamic_adjustment.available=true`；false 或 unknown 需要其他存储方案，不自动重试。扩容只有 `"status": "ready"` 才可继续，inspect 的 complete 只是证据状态，PATCH 受理不等于 Ready。扩容不能缩盘。这是 Agent 工作流，不是 CLI 自动检查，也不适用于 `bohr image build --dockerfile`。公式和停止条件见 [镜像参考](references/images.md) 与 [存储参考](references/sandbox/storage.md)。

<a id="prerequisites"></a>

## 前置条件

```bash
export BOHR_ACCESS_KEY=<your-access-key>
bohr sandbox doctor -o json   # verify local auth and control-plane connectivity
```

测试已有沙箱和已授权 HTTP(S) 目标的数据面时，可添加 `--sandbox <id> --target <url>`。此操作会连接实例，可能延长其生命周期。分层诊断和提交阶段语义见生命周期参考。

判定字段是 `data.ready`，不能只看退出码：只要诊断流程完成，`sandbox doctor` 就退出 0，不论 ready 与否。创建前检查 `data.ready`，并通过 `data.checks[]` 识别失败项。

<a id="billable-commands-need---yes-when-you-are-not-a-human-at-a-terminal"></a>

## 非交互计费命令需要 `--yes`

`sandbox create`、`image build`、`image commit` 会创建计费资源。自 2.5.18 起，CLI 在非交互终端下需要确认；Agent 的管道 stdin 默认就是非交互环境：

```json
{"ok": false, "error": {"code": "CONFIRMATION_REQUIRED", "subtype": "billing_resource",
 "hint": {"action": "confirm", "command": "bohr sandbox create --yes"}}}
```

退出码是 **10**，操作没有执行，没有创建沙箱，也没有扣费。这不是临时故障，原样重复调用仍会被拒绝。

- 费用已获授权时，在原命令上添加 `--yes` 重试。
- 费用未获授权时，报告拒绝并等待用户决定，不能自行加 `--yes`。
- 保持确认机制启用。只有用户明确要求修改持久的计费确认设置时才调整它，不能用作绕过拒绝的办法。破坏性操作仍有独立确认。
- `--dry-run` 不受计费门禁限制。价格通过 `bohr billing pricing` 查询。
- 创建前可用 `bohr sandbox create <template> --estimate-cost -o json` 预览成本：只读查询模板 SKU 价格，增加两次查询，并在 dry-run note 中显示每小时价格和有效生命周期的估算总价。默认有效期 12 小时；`--timeout 0` 和 `--never-timeout` 没有有限总价。价格不可得时会说明原因。即使指定 `--payment-source photon`，此估算仍为现金价格，不是光子扣费值。

<a id="default-sandbox-templates"></a>

## 默认模板与资源快捷入口

`bohr sandbox create` 没有模板参数，也没有 `--image`/`--gpu` 时，使用 `sdbxagent` 模板。镜像内带 Aliyun PyPI 源，国内 `pip install` 可直接使用。出站 HTTP 代理默认关闭，海外访问按 [网络参考](references/sandbox/network.md) 的 proxy on/off 步骤按需处理。

覆盖镜像时省略位置模板，可选 `--cpu <profile>`。省略 `--cpu` 时选择 `1c2g`，当前对应 `dynamic-image-replace`，资源为 1 CPU / 2 GiB，无 GPU。`--cpu` 依赖 `--image`，与位置模板/`--template` 及 `--gpu` 互斥。profile 选择 CPU/内存，`--image` 仍只替换容器镜像。创建前报告 profile、当前模板和预期资源；成功后报告响应中的真实 `templateID`、`cpuCount`、`memoryMB`。见 [CPU profile 映射](references/platform-snapshot.md#image-override-cpu-profiles)。

GPU 使用快捷参数 `--gpu`：不带值时为默认 `4090`，`--gpu 5090` 和 `--gpu l20` 选择另外两类模板。它与位置模板及 `--cpu` 互斥。

<a id="mount-a-wenyon-dataset-in-an-instance"></a>

## Wenyon 实例挂载

使用普通计算模板。数据集、挂载路径、模式和起始版本都是 **sandbox create 参数**，不是模板默认值。参数见 `bohr sandbox create --help`，完整流程见 [生命周期参考](references/sandbox/lifecycle.md#mount-a-wenyon-dataset-in-an-instance)。

```bash
REQUEST_ID="$(uuidgen)"

bohr sandbox create sdbxagent \
  --wenyon-dataset-id <dataset-id> \
  --wenyon-mount-path /mnt/wenyon \
  --wenyon-mode read --wenyon-version current \
  --wenyon-user-token-env WENYON_USER_JWT \
  --request-id "$REQUEST_ID" --yes -o json
```

对本次逻辑 create 只生成一次 `REQUEST_ID`，在不同 shell 或工具调用之间保留它。核对和恢复原请求时复用该 ID，不要为了重试重新生成。

默认 `read/current` 在创建租约时固定当前版本。`write/current` 或正整数版本会完整预载该基线，只有 `write/none` 从空目录开始。版本不存在或预载失败时，不会回退为空目录。检查响应中的数据集、路径、请求和解析后的版本及实际读写模式，并核对挂载内容。

`--wenyon-user-token-env` 接受持有 Wenyon/Vouch 用户 JWT 的现有环境变量名，不接受 token 字面量。AccessKey 鉴权和 `--as user` 均不提供该 JWT。`bohr auth token` 报 `auth_method=access_key` 时不能复制其输出作为 JWT。token 不能进入命令参数、元数据、dry-run、日志或对话。不能与 `--inherit-auth` 组合。

可写挂载必须显式选择结算：`sandbox delete --commit --wait` 发布整个工作区，包括删除；`--discard --wait` 丢弃工作区。`--force` 只跳过进程检查，不选择结算方式。commit 后核对 `produced_version`。未使用 `--wait` 时，受理不代表数据已保存。

可写执行失败或过期且工作区仍可用时，检查 `recovery_deadline`，在一小时 `awaiting_decision` 窗口内选择结算。期间持有写锁，截止后工作区被丢弃；上游驱逐可能更早使恢复不可用。恢复同一 create 期间保持原 request ID，不要重复创建。沙箱有效期不是 HTTP 或挂载就绪超时，镜像 `READY` 也不能证明数据集就绪。

<a id="command-surface"></a>

## 命令入口

| 命令 | 用途 |
| --- | --- |
| `bohr sandbox` | 操作云沙箱 |
| `bohr image` | 镜像目录、沙箱快照与构建 |

<a id="sandbox-commands"></a>

## 沙箱命令

| 子命令 | 说明 |
| --- | --- |
| `doctor` | 检查配置与就绪状态 |
| `quota` | 配额上限和用量，创建前或 limit reached 后查询 |
| `storage inspect <id>` | 只读查询 rootfs 容量、用量和可用空间 |
| `storage resize <id> --target <N>Gi` | 扩容运行中沙箱并等待可验证的 Ready |
| `create [template]`（`run`） | 创建沙箱，默认 sdbxagent；`--parent-sandbox-id` 记录父沙箱所有权 |
| `list`（`ls`） | human 输出默认显示父子树，`--flat` 改为平铺；结构化/quiet 输出始终平铺；`--query failed` 诊断近期失败 |
| `describe <id>`（`get`） | 只读详情，不唤醒暂停实例 |
| `connect <id>` | 解析数据面连接 |
| `exec <id> <cmd>` | 在沙箱内以 root 运行命令，支持前台和后台 |
| `ps <id> [--pid <pid>]` | 列出进程，或按精确 PID 读取完整既有字段 |
| `process connect <id> <pid>` | 仅观察本次精确 PID 连接后收到的输出 |
| `process kill <id> <pid>` | 解析精确 PID 并确认后请求 SIGKILL，只报告 RPC 受理 |
| `delete <id>`（`kill`） | 终止沙箱；删除父节点会级联其已记录的子孙节点 |
| `files write --source <src> <id> <path>` | 上传文件或目录 |
| `files read <id> <path> [--destination <dst>]` | 读取或下载文件 |
| `terminal attach <id>` | 将本地 TTY 附着到 PTY，类似 docker exec -ti |
| `terminal create <id>` | 以 root 创建 PTY |
| `terminal send <id> <pid> <input>` | 向 PTY 发送输入 |
| `terminal kill <id> <pid>` | 终止 PTY |
| `machine list`（`machine ls`） | 列出机器 SKU |
| `template create` | 创建模板，`--visibility 1` 公开、`2` 私有，默认私有 |
| `template list`（`template ls`） | 列出模板 |
| `template update` | 更新模板镜像 |
| `template delete`（`template rm`） | 删除模板 |
| `template share` | 将私有模板授权给用户或项目主体 |
| `template unshare` | 撤销具体用户或项目授权 |
| `template grants` | 列出私有模板的用户和项目授权 |

进程选择器只接受精确的正 uint32 PID。`process connect` 只保证收到附着流后的输出，不保证历史回放或完整日志。观察到结束事件时传播其退出码；缺少结束事件时返回 `sandbox_process_outcome_unknown`，`retryable=false`。

`process kill` 在 List 预查和延迟确认后发送 SIGKILL。`signal_accepted` 仅表示 SendSignal 已受理，`termination_confirmed` 始终为 false。发送信号后的传输失败属于结果未知，不能盲重试。List 和操作之间可能发生 PID 复用，数据面凭据能作用于该沙箱中任何可见 PID。使用 `ps --pid` 再次核对；PTY 使用 `terminal attach` 或 `terminal kill`，需要优雅退出的任务使用手工 TERM 流程。

<a id="discovering-commands"></a>

## 镜像命令

旧 `bohr sandbox image` 入口已移除，使用独立镜像领域的完整路径：

| 命令 | 用途 |
|---|---|
| `bohr image commit` | 为运行中沙箱制作快照，先做根盘检查 |
| `bohr image build` | 从 Dockerfile 构建，选择本地或远端模式 |
| `bohr image build-log <id>` | 读取或跟随构建日志 |
| `bohr image dockerfile <id>` | 读取保存的 Dockerfile 原文 |
| `bohr image get <id>` | 读取镜像目录记录 |
| `bohr image list` | 读取一页镜像目录 |

## 发现命令

```bash
bohr sandbox --help
bohr sandbox <subcommand> --help
```

<a id="version-contract"></a>

## 版本契约

源 Skill 随 `bohr` 二进制内嵌并与所描述命令保持相同版本契约。本公开包的来源版本见仓库的 `cli-skills.lock.json`。可读取已安装 CLI 的对应版本：

```bash
bohr guide read bohr-sandbox
bohr guide list bohr-sandbox/references
```

`bohr guide` 不访问云端 Skill Hub。升级 `bohr` 可获得新版命令指引；本地包的完整参考文档已经随 Skill 分发，安装后可直接读取。

<a id="reference-docs"></a>

## 参考文档

详细文档位于本地 `references/`，按任务需要读取。

**沙箱生命周期和使用：**

- [默认行为陷阱](references/sandbox/pitfalls.md)：30Gi 基础盘、创建时额外磁盘、配额、超时、Wenyon 就绪/JWT、二进制下载、删除、代理、计费和 ML 训练检查表。
- [生命周期](references/sandbox/lifecycle.md)：create、list、describe、delete、quota 和实践规则。
- [执行模式](references/sandbox/execution-modes.md)：前台、后台、PTY，以及长任务先取回结果再删除的 SOP。
- [exec](references/sandbox/exec.md)：命令执行，包括 `--background`。
- [终端](references/sandbox/terminal.md)：用于 REPL、TUI 和 Ctrl-C 的 PTY。
- [文件](references/sandbox/files.md)：文件和目录上传下载。
- [远程存储索引](references/sandbox/remote-storage-index.md)：通过私有、会过期的路径索引避免反复远程扫描。
- [存储](references/sandbox/storage.md)：rootfs 证据、运行时扩容和 image commit 前检查。
- [模板](references/sandbox/templates.md)：模板查询、创建、更新、删除和接口边界。
- [网络](references/sandbox/network.md)：按环境批准的代理配置临时开启和关闭出站 HTTP 代理。

**容器镜像：**

- [镜像参考](references/images.md)：commit/build、状态和日志，以及恢复已保存的 Dockerfile。

**沙箱内安装软件：**

- [软件安装](references/software-install/README.md)：预装软件清单和国内镜像源。

**平台快照：**

- [平台快照](references/platform-snapshot.md)：固定版本的 CPU/GPU 模板快捷入口和 SKU 发现。

<a id="output-formats"></a>

## 输出格式

`bohr sandbox` 支持通用 `-o json|human|table|ndjson|csv|yaml` 输出。脚本和 Agent 使用 `-o json`，可选 `-q <jq expression>` 过滤结果。

<a id="resource-payment-boundary"></a>

## 资源支付边界

标记为 balance-only 的资源操作继续按原有规则从余额计费。全局 `--payment-source` 不会为这些资源启用光子支付；显式指定 `photon` 会在执行前被拒绝。按该资源既有的价格查询和费用授权流程操作。
