---
name: bohrium-knowledge-base
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr kb 创建、查询、搜索或删除知识库，上传文献、列出和读取文档或删除文档；用于知识库、父节点和文档 ID 的正确交接。"
---

<a id="bohrium-knowledge-base"></a>

# Bohrium 知识库

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 搜索已有私有内容 | 确定 KB ID → 定向 `search` → 检查返回证据 |
| 添加本地文档 | 确认文件及 KB → 上传 → 定位稳定文档 ID → 检查索引 → 验证检索 |
| 删除内容 | list 或 get 确认文档或 KB → 获授权删除；文件夹 ID 单独处理 |

## 工作流：上传论文并验证能够检索

1. 从 `kb list` 确定目标知识库，确认本地文件和目的地后上传。按下方选择器规则使用 `--kb-id` 或 `--parent-id` 之一。

```bash
bohr kb upload ./paper.pdf --kb-id <kb_id> -o json
bohr kb document list --kb-id <kb_id> --page 1 --size 50 -o json
bohr kb document get <document_id> -o json
bohr kb search <kb_id> "distinctive phrase from the uploaded paper" -o json
```

2. 保留上传返回的 `data.knowledge_base_id`、`parent_id` 和 `md5`。`data.document.index_state=submitted` 只表示注册成功，不表示索引完成。上传后回查尚未找到文档时，`data.document.document_id` 可能缺失；用文档列表定位，不要重新上传。
3. 在 `data.documents[]` 匹配目标文档，保存其 `document_id` 供 get 或 delete 使用。检查该行 `index_state`：`processing` 可有界复查，`ready` 可进行检索验证，`failed` 或 `unsupported` 则报告失败或处理格式。`unavailable` 或 `unknown` 表示缺少证据，不能无限轮询。页级 `index_status_available=false` 也可能由其他行导致。
4. `document get` 检查安全元数据，不读取全文。搜索有辨识度的内容，确认真实命中来自目标文档后，才能声称可检索。显式 `recallable=false` 的限制要保留；单次搜索无命中不能证明索引失败。
5. 交付 KB ID、稳定文档 ID、观察到的索引状态及检索证据。注册成功但索引或检索未核实时，准确报告该部分结果。继续检查原文档；重复上传属于新写入，不是就绪探针。

先确定知识库 ID：

```bash
bohr kb list --page 1 --size 20 -o json
bohr kb create "Project literature" -o json
```

知识库 ID 标识 KB 本身。上传和列出文档在内部使用根节点或父节点 ID；CLI 可根据 `--kb-id` 解析根节点：

```bash
bohr kb upload ./paper.pdf --kb-id <kb_id> -o json
bohr kb document list --kb-id <kb_id> --page 1 --size 50 -o json
```

用户选中了具体根节点或文件夹节点时，使用 `--parent-id`。父节点 ID 不等于知识库 ID，也不等于文档 ID。

在单个知识库内搜索内容：

```bash
bohr kb search <kb_id> "catalyst deactivation mechanism" -o json
```

先从 `document list` 获取稳定文档 ID，再读取或删除：

```bash
bohr kb document get <document_id> -o json
```

upload 会将指定本地文件传到 Bohrium。写入前确认文件、目标知识库或父节点，以及显示名称。当前 KB 命令均免费，但免费不代表允许上传或删除。

知识库和文档删除均属于破坏性操作。添加 `--yes` 前立即核对具体目标：

```bash
bohr kb document delete <document_id> --yes -o json
bohr kb delete <kb_id> --yes -o json
```

当前 CLI 不管理文件夹、标签、笔记、收藏、分享或权限。完整命令、标识符边界、上传路由和删除确认见 [命令参考](references/commands.md)。
