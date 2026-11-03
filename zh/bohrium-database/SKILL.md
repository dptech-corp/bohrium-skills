---
name: bohrium-database
compatibility: "需要 bohr CLI（npm 包 @dptech-corp/bohr-cli）；本 Skill 按 2.9.19 核验。账户操作需要 Bohrium 身份和网络访问。"
description: "通过 bohr database 查询聚合物或结构数据库，以及使用 db_ak/table_ak 发现私有表、读取 schema 和查询数据；用于科学记录检索和结构文件下载。"
---

<a id="bohrium-database"></a>

# Bohrium 科学数据库

## 按任务选择路径

| 目标 | 路径 |
|---|---|
| 获取晶体结构 | 搜索 → 检查可用结构及警告 → 按需下载归档 → 验证本地输入 |
| 浏览高分子 | 限定数量列出 → 查看科学记录 |
| 查询已有私有表 | 已提供 `db_ak` → tables → schema → 按类型构造过滤和查询，前提是路由已启用 |

## 工作流：取得可用于计算的结构

1. 明确成分或查询、最大数量、CIF 或 POSCAR 格式，以及本地目标。远端只读，但下载归档会写入本地。

```bash
bohr database structure search --query "Li Fe P O" --n-results 5   --output-format cif --download --out ./structures.tgz -o json
```

2. 从 `data.cleaned_structures` 查看可用科学记录。命令使用同次搜索的 `data.output_dir` 下载归档；它是取回数据的 URL，不是私有库 `db_ak` 或本地路径。
3. 检查 `notice.warnings` 和 `data.errors`。`STRUCTURE_SEARCH_PARTIAL` 表示调用成功但部分来源缺失。把本地文件选为 Job 输入前，检查归档内容和结构格式；命中数量或下载文件名不能证明计算输入有效。
4. 交付所选结构的出处、缺失来源说明及实际核实的本地文件。下载失败时保留搜索记录，按参考中的恢复提示和目标路径规则处理。不可用的归档链接不能靠重复下载修复。
5. 私有表先确认路由可用：正式 CLI 可能返回 `db_route_unavailable`，且没有私有库发现命令。路由启用时，将表记录选中的 `table_ak` 传给 schema 和 query，再按 schema 的列名及类型构造过滤值。路由失败不等于空数据库，也不能靠编造键修复。

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
