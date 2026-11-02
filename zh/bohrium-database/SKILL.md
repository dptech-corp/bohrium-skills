---
name: bohrium-database
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr database 查询聚合物或结构数据库，以及使用 db_ak/table_ak 发现私有表、读取 schema 和查询数据；用于科学记录检索和结构文件下载。"
---

<a id="bohrium-database"></a>

# Bohrium 科学数据库

先选择数据库类型：

- 领域库使用 `bohr database polymer` 或 `bohr database structure`，返回科学记录，不使用私有表标识。
- 私有表格访问使用 `tables`、`schema` 和 `query`，传入稳定的 `db_ak` 和 `table_ak`。

当前数据库命令均为只读且免费，优先使用结构化输出：

```bash
bohr auth whoami -o json
bohr database polymer list --page 1 --size 20 -o json
bohr database structure search --query "Si O" --n-results 5 -o json
```

私有数据库先发现表，再查询行：

```bash
bohr database tables <db_ak> -o json
bohr database schema <db_ak> <table_ak> -o json
bohr database query <db_ak> <table_ak> \
  --filter '[{"field":"status","operator":"eq","value":"active"}]' \
  --limit 20 \
  --offset 0 \
  -o json
```

显示名称是标签，不是标识符。在发现、schema 和 query 调用之间原样保留 `db_ak` 与 `table_ak`。filter 是 JSON 数组，应使用返回 schema 中的字段名和类型，不能猜测。

结构搜索可返回 CIF 或 POSCAR 输出元数据，并可选下载结果压缩包。使用 `--download` 前确认本地目标位置；远程操作仍只读，但会写入本地文件。

当前 CLI 不能列出私有数据库，也不能修改行、表或列。数据库创建、插入、更新、删除和 schema 变更均不在当前命令范围内，应如实报告能力边界。

完整命令、参数、标识符和输出/下载边界见 [命令参考](references/commands.md)。
