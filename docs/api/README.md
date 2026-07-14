# Bohrium OpenAPI — 单一事实源（SSOT）

本目录维护 Bohrium 平台对外 API 文档的**唯一事实源**，内容基于各 Skill（`zh/<skill>/SKILL.md`）中记录的真实接口。新增/修改接口或调整定价，都改这里；Apifox 定时从本仓库抓取 `openapi.json` 并同步到 <https://open.bohrium.com>，飞轮自动运转，无需手工维护 Apifox。

## 文件

| 文件 | 说明 |
|------|------|
| `openapi.json` | canonical spec（OpenAPI 3.0.1）。Apifox 的定时数据源指向此文件。改动接口/定价改它。 |

## 规范

- **鉴权**：所有接口使用 `Authorization: Bearer <BOHR_ACCESS_KEY>`（`components.securitySchemes.bearerAuth`）。
- **服务地址**：`https://open.bohrium.com`。
- **返回信封**：`{code, data, message}`，`code==0` 成功，`code==2000` 未授权。
- **分组**：每个 Skill 对应一个 tag，命名 `中文名 (bohrium-xxx)`。
- **定价**：计费 operation 加 `x-bohrium-price` 扩展，并在 `description` 追加一行 `**计费**：…`。

### `x-bohrium-price`

```json
"x-bohrium-price": {
  "billable": true,
  "currency": "CNY",
  "items": [{ "condition": "type=0", "amount": 0.05, "unit": "次" }],
  "skill": "bohrium-paper-search",
  "note": "查询结果免费"
}
```

- `items[].condition` 可选，用于按参数区分档位（如 `type=0`）；`unit` 常见 `次`/`页`/`小时`。
- 与根目录 `README.md` 的「计费说明」表保持一致。

## 维护流程

1. 改动代码或 `zh|en/*/SKILL.md` 中的接口/定价时，同步更新 `openapi.json` 对应 operation。
2. 校验：`python3 -c "import json;json.load(open('docs/api/openapi.json'))"`，并可用 `openapi-spec-validator` 或 `npx @redocly/cli lint` 做 3.0 合规检查。
3. 合并到 `main` 后，Apifox 在下一个抓取周期（每 3 小时）自动拉取。

## Apifox 数据源配置（一次性）

项目设置 → 数据管理 → 绑定数据源：

- 目标分支：`main`
- 数据源格式：`OpenAPI (Swagger)`
- 导入频率：`每隔 3 小时`
- 数据源：`Git 仓库` 连接 `dptech-corp/bohrium-skills`，文件路径 `docs/api/openapi.json`（或用该文件 raw URL + Basic Auth）
- 导入模式：`智能合并`

## 认责与追溯

底层服务由各团队开发，本仓库集中维护契约文档。认责链路：

- **契约归属**：每个模块的接口契约以 `zh|en/<skill>/SKILL.md` 为准（`openapi.json` 由其派生）。`.github/CODEOWNERS` 将触及某模块 SKILL.md 的 PR 自动指派给该团队 review —— 谁改契约、谁负责就此绑定。
- **模块注册表**：`docs/api/owners.yaml` 记录每个模块的 `path_prefixes` 与 `team/owner/contact/service_repo/oncall`。各团队认领即填入自己的信息。
- **从日志定位责任人**：`python3 tools/whoowns.py "<路径或日志行>"`，按最长前缀匹配得出模块、团队、联系方式与契约文档；`--json` 输出便于 Agent 调用。
- **改动审计**：`git blame zh/<skill>/SKILL.md` 查最近改动人；PR #编号 追溯到具体变更与讨论。
- **规范内嵌**：`openapi.json` 每个 tag 上带 `x-owner`（由 owners.yaml 注入），Agent 读 spec 即可知道归属。

运维流程示例：拿到日志 → `whoowns.py` 定位模块与团队 → 按 owners.yaml 联系方式派单 → 团队改 SKILL.md → PR 经 CODEOWNERS review 合并 → Apifox 自动同步。
