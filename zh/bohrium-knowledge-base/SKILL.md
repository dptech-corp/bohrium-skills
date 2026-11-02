---
name: bohrium-knowledge-base
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr kb 创建、查询、搜索或删除知识库，上传文献、列出和读取文档或删除文档；用于知识库、父节点和文档 ID 的正确交接。"
---

<a id="bohrium-knowledge-base"></a>

# Bohrium 知识库

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
