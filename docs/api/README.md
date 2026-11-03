# Bohrium OpenAPI — 单一事实源（SSOT）

本目录独立维护 Bohrium 平台对外 API 规范。CLI 版 Skills 通过 `bohr guide` 读取指引，不再记录接口定义或维护本目录的副本。新增/修改接口或调整 API 定价时，更新本目录规范；Apifox 定时从本仓库抓取 `openapi.json` 并同步到 <https://open.bohrium.com>，飞轮自动运转，无需手工维护 Apifox。

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
- API 定价由本规范独立维护；CLI 用户通过 `bohr billing pricing` 和当前内嵌指引核对价格。

## 维护流程

1. API 实现或定价发生变更时，同步更新 `openapi.json` 对应 operation；Skill 文字更新不作为接口变更来源。
2. 校验：`python3 -c "import json;json.load(open('docs/api/openapi.json'))"`，并可用 `openapi-spec-validator` 或 `npx @redocly/cli lint` 做 3.0 合规检查。
3. 合并到 `main` 后，Apifox 在下一个抓取周期（每 3 小时）自动拉取。

## Apifox 数据源配置（一次性）

项目设置 → 数据管理 → 绑定数据源：

- 目标分支：`main`
- 数据源格式：`OpenAPI (Swagger)`
- 导入频率：`每隔 3 小时`
- 数据源：`Git 仓库` 连接 `dptech-corp/bohrium-skills`，文件路径 `docs/api/openapi.json`（或用该文件 raw URL + Basic Auth）
- 导入模式：`智能合并`
