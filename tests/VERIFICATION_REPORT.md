> **历史 API 验证记录。** 本文记录旧版直接调用 API 的 Skills，不代表当前 CLI 版的验收结果。配套 API 冒烟和示例提取脚本已随 CLI 迁移退役；当前检查见 `.github/workflows/check.yml`。

# Bohrium Skills API 端点验证报告

**最近验证**: 2026-06-30（bohrium-file 全链路实测 + 文档/校验补齐）｜2026-06-26（bohrium-sciencepedia 重写后 12 端点真实实测，原 bohrium-wiki 重命名）｜2026-06-15（全量真实冒烟，含 mentor/sandbox）｜2026-06-11（bohrium-tools 真实实测）｜2026-06-08（v2 网关冒烟实测）｜详细功能基线: 2026-05-11
**测试 AK**: 通过 `BOHR_ACCESS_KEY` 环境变量注入（不在报告中明文记录）
**API Base**: https://open.bohrium.com/openapi （多数 skill 使用 v2；`bohrium-job` 和 `bohrium-file` 的常规文件操作保留 v1）
**bohr CLI**: v1.1.0 (Go, 从 OSS 安装到 ~/.bohrium/bohr)
**lbg CLI**: v4.0.0b47 (Python >=3.10 prerelease, pip install --pre lbg)
**OPENAPI_HOST**: https://open.bohrium.com

> 网关版本：多数 skill 的 OpenAPI 网关路径已统一升级到 `/openapi/v2/*`，上游服务与 v1 一致（多数仅外层版本号变化）。例外：`bohrium-job` 保留 v1（v2 上游不同且 `job_group` 在网关无 v2 路由）；`bohrium-file` 的 personal/share 常规文件操作保留 v1，只有上传凭证使用 v2。镜像 / 数据集 / 项目等域名已统一为 `open.bohrium.com`（历史上 dataset create 需走 `openapi.dp.tech` 的 307 问题已在 open-platform 修复）。

---

## 网关冒烟实测（2026-06-15 全量 + 2026-06-30 bohrium-file 补测，`tests/smoke_test.py`）

每个 skill 打一个主端点，真实请求 `open.bohrium.com`（非 mock）。结果：**PASS=17, FAIL=0, SKIP=0**（`mentor` 创建会话，`sandbox` 创建短时沙箱并执行命令后销毁；`bohrium-file` 使用 v1 常规文件接口 + v2 上传凭证接口）。

| Skill | 端点 | 方法 | 结果 | 计费 |
|-------|------|------|------|------|
| bohrium-job | `/v1/job/list` | GET | ✅ PASS | 免费（保持 v1） |
| bohrium-node | `/v2/node/list` | GET | ✅ PASS | 免费 |
| bohrium-dataset | `/v2/ds/` | GET | ✅ PASS | 免费 |
| bohrium-file | `/v1/file/iterate`、`/v1/file/stat/*`、`/v2/file/upload/binary` | POST/GET | ✅ PASS | 免费（上传凭证；未上传 bytes） |
| bohrium-image | `/v2/image/public` | GET | ✅ PASS | 免费 |
| bohrium-project | `/v2/project/lite_list` | GET | ✅ PASS | 免费 |
| bohrium-knowledge-base | `/v2/knowledge/knowledge_base/list` | GET | ✅ PASS | 免费 |
| bohrium-paper-search | `/v2/paper/rag/pass/keyword` | POST | ✅ PASS | 余额扣费 |
| bohrium-paper-search | `/v2/paper/rag/pass/patent` | POST | ✅ PASS | 余额扣费 |
| bohrium-pdf-parser | `/v2/parse/trigger-url-async` | POST | ✅ PASS | 按页余额扣费 |
| bohrium-web-search | `/v2/search/web` | GET | ✅ PASS | 免费（v2 取消 v1 限额） |
| bohrium-scholar-search | `/v2/paper-server/scholar/search` | POST | ✅ PASS | 免费 |
| bohrium-sciencepedia | `/wiki_v2/{info, search/universal, article, major_levels, level_fields, get_wiki_index, knowledge_graph, search_index_name}` | GET / POST | ✅ PASS ×8（任务链路冒烟） | 免费 |
| bohrium-tools | `/v2/literature-sage/tool/domain` | GET | ✅ PASS | 免费 |
| bohrium-tools | `/v2/literature-sage/tool/search/hybrid` | POST | ✅ PASS | 免费 |
| bohrium-mentor | `/v2/sigma-search/api/v4/ai_search/sessions` | POST | ✅ PASS | 创建会话扣余额 |
| bohrium-sandbox | `lbg sdbx create/exec/kill`（launching/v2） | CLI | ✅ PASS | 需 lbg beta；create/exec 计费 |

> 计费原则：原本无计费的列表/查询类端点 v2 后仍免费；原本即计费的（paper-search 余额、pdf-parser 限额→按页余额）照常计费；`mentor` 创建会话与 `sandbox` create/exec 会扣费，当前冒烟会真实执行。

---

## 总览

| 状态 | Skill | 说明 |
|------|-------|------|
| ✅ 完全可用 | bohrium-project, bohrium-pdf-parser, bohrium-web-search, bohrium-sandbox, bohrium-job, bohrium-node, bohrium-knowledge-base, bohrium-image, bohrium-scholar-search, bohrium-sciencepedia, bohrium-tools, bohrium-lkm, bohrium-paper-search, bohrium-dataset, bohrium-database, bohrium-file, bohrium-mentor | 当前仓库 17 个 skill，文档端点 / CLI 均正常 |
| ❌ 已移除 | polymer-db, bohrium-viking-memory, bohrium-scholar, bohrium-matmaster, diagnose-agent, proposal-agent, preparation-agent, scoring-agent | 已下架 / 冗余 / 后端不可用；当前仓库不含这些 skill |

---

## 逐 Skill 详细结果

### bohrium-job (任务管理) — 保留 v1

**推荐方式**: 使用 `bohr` CLI（Go 版本，从 OSS 安装）

> `bohrium-job` 不升级 v2：网关 `/openapi/v2/job` 指向不同上游（`/brm/v2/job`），且 `job_group` 在网关没有 v2 注册，升级会 404。保持 `/openapi/v1/*`。

#### CLI 测试 (bohr v1.1.0)

| 命令 | 状态 | 备注 |
|------|------|------|
| `bohr job list -n 5 --json` | ✅ | 返回 id/jobName/status/cost |
| `bohr job list -r` / `-f` / `-i` / `-p` | ✅ | 按状态过滤正常 |
| `bohr job describe -j {id} --json` | ✅ | 完整详情含 cmd/imageName/machineType |
| `bohr job log -j {id}` | ✅ | 自动下载 log 文件到本地 |
| `bohr job download -j {id} -o ./` | ✅ | 下载任务结果 |
| `bohr job submit -m ... -t ... -c ... -p ...` | ✅ | 提交成功，返回 JobId + JobGroupId |
| `bohr job kill {id}` | ✅ | 强制停止 |
| `bohr job terminate {id}` | ⚠️ | Pending 状态任务不可 terminate |
| `bohr job_group list -n 5 --json` | ✅ | 正常 |
| `bohr job_group create -n ... -p ...` | ✅ | 创建成功返回 job_group_id |
| `bohr job_group terminate {id}` / `delete {id}` | ✅ | 正常 |

#### API 补充测试

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v1/job/list?status=N` | GET | ✅ | 按状态过滤正常（冒烟实测 PASS） |
| `/v1/job/view/conf/{job_id}` | GET | ✅ | 返回 state/token/host |
| `/v1/job/{job_id}/modify` | POST | ✅ | 重命名成功 |
| `/v1/job_group/{id}/modify` | POST | ✅ | 任务组重命名成功 |

**结论**: CLI 全功能可用。`POST /job/submit`、`GET /job_group/list` 需通过 CLI 完成（API 直连 404）。

---

### bohrium-node (开发机)

#### CLI 测试 (bohr v1.1.0)

| 命令 | 状态 | 备注 |
|------|------|------|
| `bohr node list --json` | ✅ | 返回 nodeId/nodeName/status/ip |
| `bohr node stop {nodeId}` | ✅ | 停止成功 |
| `bohr node create` | ⚠️ | 交互式（需 TTY），自动化需用 API |

#### API 补充测试（v2，2026-06-08 实测）

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/node/list` | GET | ✅ | 节点列表（冒烟实测 PASS） |
| `/v2/node/resources` | GET | ✅ | CPU/GPU/磁盘选项 |
| `/v2/node/resources/price?skuId=&projectId=` | GET | ✅ | 返回价格（元/小时） |
| `/v2/node/{nodeId}` | GET | ✅ | 详情含 IP/状态/密码 |
| `/v2/node/add` | POST | ✅ | 创建节点（资源不足时正常报错） |
| `/v2/node/restart/{nodeId}` | POST | ✅ | 端点存在（不存在的 id 返回 record not found） |
| `/v2/node/modify/{nodeId}` | POST | ✅ | 端点存在（同上） |
| `/v2/node/ds` / `/v2/node/ds/bind` | GET/POST | ✅ | 端点存在 |
| `/v2/node/start/{nodeId}` | POST | ❌ 404 | 启动端点不存在于网关（文档用 restart 而非 start） |

**结论**: list/resources/add/restart/modify/ds 等全部可用；仅 `start/{id}` 在网关无路由，节点文档使用 `restart`。

---

### bohrium-dataset (数据集)

**推荐方式**: 使用 `bohr` CLI（Go 版本）

> 域名已统一 `open.bohrium.com`：历史上 `POST /openapi/v1/ds/` 因尾斜杠触发 307 重定向、create 需绕道 `openapi.dp.tech` 的问题，已在 open-platform 修复（tag b_open-platform_2.0.7）。现 list / create / delete 均直接走 `open.bohrium.com`。

#### CLI 测试 (bohr v1.1.0)

| 命令 | 状态 | OPENAPI_HOST | 备注 |
|------|------|------|------|
| `bohr dataset list --json` | ✅ | open.bohrium.com | 返回 id/title/path/projectName |
| `bohr dataset list -t {title}` / `-p {projectId}` | ✅ | open.bohrium.com | 按标题/项目过滤 |
| `bohr dataset create -n ... -p ... -i ... -l ...` | ✅ | open.bohrium.com | 创建+上传成功 |
| `bohr dataset delete {id}` | ✅ | open.bohrium.com | 删除成功 |

#### API 补充测试

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/ds/?keyword=&pageSize=&pageNum=` | GET | ✅ | 列表+搜索（冒烟实测 PASS） |
| `/v2/ds/{id}` | GET | ✅ | 详情 |
| `/v2/ds/` | POST | ✅ | 创建（307 问题已修复） |

**结论**: CLI 和 API 全功能可用，统一使用 `open.bohrium.com`。

---

### bohrium-file (文件盘)

**推荐方式**: 常规 personal/share 文件操作使用 `/openapi/v1/file/*`；上传优先使用 `/openapi/v2/file/upload/binary` 获取 storage/NAS 上传凭证后再上传 bytes。

#### API 实测（2026-06-30）

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v1/ak/get` | GET | ✅ | 获取当前 `user_id` / `orgId` |
| `/v1/file/iterate` | POST | ✅ | personal 盘列目录；字段为 `prefix` |
| `/v1/file/stat/*` / `/v1/file/meta/*` | GET | ✅ | personal 路径存在性和元数据 |
| `/v1/file/download/*` | GET | ✅ | 下载需跟随重定向 |
| `/v1/file/mkdir` | POST | ✅ | 临时目录创建成功，已清理 |
| `/v1/file/copy` / `/copyr` | POST | ✅ | 文件复制、目录递归复制成功，已清理 |
| `/v1/file/move` / `/mover` | POST | ✅ | 文件移动、目录递归移动成功，已清理 |
| `/v1/file/delete/*` / `/deleter/*` | DELETE | ✅ | 文件删除、目录递归删除成功 |
| `/v1/file/decompress` | POST | ✅ | zip 解压成功，已清理 |
| `/v1/file/search/config` / `/recent` / `/transfer/list` | POST/GET | ✅ | 查询类端点正常 |
| `/v1/file/upload/binary` | POST | ✅ | v1 直接上传返回 307；`curl -L` 后文件上传成功 |
| `/v2/file/upload/binary` | GET | ✅ | 返回 `host`、`Authorization`、`X-Storage-Param`；未上传 bytes 的冒烟也通过 |
| `/v2/file/modify` | POST | ✅ | 修改历史上报接口正常 |
| `/v2/file/iterate` with `pathType=personal` | POST | ✅ | 按预期返回 `path type not found`，说明 personal/share 不走 v2 iterate |

**鉴权补充**: 生产网关实测 `Authorization: Bearer <AK>` 与 `accessKey: <AK>` 均可用；`BOHR_ACCESS_KEY: <AK>` header 返回 401（入口层可能过滤下划线 header）。skill 文档统一使用环境变量 `BOHR_ACCESS_KEY` + HTTP header `Authorization: Bearer $BOHR_ACCESS_KEY`。

**结论**: personal/share 文件盘应按 v1 使用；v2 的 `upload/binary` 是 storage/NAS 上传凭证签发，不是裸 OSS 地址。最终落盘位置由第一步的 `path=/personal/...` 或 `path=/share/...` 决定。`/v2/file/iterate` / `/v2/file/download` 只适用于 `pathType=appJob`、`pathKey=appJobId`。

---

### bohrium-image (镜像)

#### CLI 测试 (bohr v1.1.0)

| 命令 | 状态 | 备注 |
|------|------|------|
| `bohr image list --json` | ✅ | 自定义镜像列表 |
| `bohr image delete {id}` | ✅（文档） | — |

#### API 测试（v2，2026-06-08 实测全部 11 端点）

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/image/public` | GET | ✅ | 公共镜像分类（冒烟实测 PASS） |
| `/v2/image/public/version/search?keyword=` | GET | ✅ | 版本搜索 |
| `/v2/image/public/{imageId}/version` | GET | ✅ | 指定镜像版本列表 |
| `/v2/image/last_used?type=` | GET | ✅ | 最近使用镜像 |
| `/v2/image/private?device=&type=` | GET | ✅ | 私有镜像列表 |
| `/v2/image/private/{id}` | GET | ✅ | 私有镜像详情 |
| `/v2/image/private` | POST | ✅ | 创建（dockerfile 需 base64） |
| `/v2/image/private/{id}` | PUT | ✅ | 修改描述（端点存在） |
| `/v2/image/dockerfile/check` | POST | ✅ | Dockerfile 合法性校验 |
| `/v2/image/{id}/share` | POST/DELETE | ✅ | 分享/取消（端点存在） |

**结论**: 所有 API 端点正常。中英文 SKILL.md 端点已对齐（last_used / 更新描述 / 私有详情等）。

---

### bohrium-project (项目管理)

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/project/lite_list` | GET | ✅ | 轻量列表（冒烟实测 PASS） |
| `/v2/project/list` | GET | ✅ | 完整列表（含费用详情） |
| `/v2/project/set_name` | POST | ✅ | 重命名成功 |
| `/v2/project/add_user` | POST | ✅ | 添加成员成功 |

**结论**: 全部端点正常。

---

### bohrium-knowledge-base (知识库)

> 网关 `/openapi/v2/knowledge/*` → `KnowledgeDBHost`（literature-sage 地址），路径由 transformer 改写。

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/knowledge/knowledge_base/list` | GET | ✅ | 知识库列表（冒烟实测 PASS） |
| `/v2/knowledge/knowledge_base/create` | POST | ✅ | 创建成功 |
| `/v2/knowledge/knowledge_base/update` | POST | ✅ | 更新成功（需 NodesId） |
| `/v2/knowledge/knowledge_base/{nodeId}` | GET | ✅ | 详情正常 |
| `/v2/knowledge/knowledge_base/search/name` | POST | ✅ | 按名称搜索 |
| `/v2/knowledge/file` / `/file/submit` / `/file/search` | GET/POST | ✅ | 文献列表/注册/内容搜索 |
| `/v2/knowledge/recall/papers` / `/recall/hybrid` | POST | ✅ | 语义/混合召回 |
| `/v2/knowledge/folder/delete` | POST | ✅ | 删除知识库/文件夹 |

**结论**: 全部端点正常。正确路径为 `/folder/delete`、`/file`（无 `/list` 后缀）、`/file/search`、`/recall/hybrid`。

---

### bohrium-paper-search (论文与专利) — 余额计费

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/paper/rag/pass/keyword` (基础/高级) | POST | ✅ | 正常（冒烟实测 PASS） |
| `/v2/paper/rag/pass/patent` (基础) | POST | ✅ | 正常（冒烟实测 PASS） |

**计费**: v2 `paper/rag` 走余额扣费（BalanceBill）。**结论**: 论文/专利搜索正常。

---

### bohrium-pdf-parser (PDF 解析) — 按页余额计费

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/parse/trigger-url-async` | POST | ✅ | 返回 token + page_count（冒烟实测 PASS） |
| `/v2/parse/get-result` (content/objects/pages_dict) | POST | ✅ | 查询结果，不计费 |

**计费**: v2 trigger 成功后按响应 `page_count` 扣余额（扣成功才下发 token）；`get-result` 等查询端点不计费。**结论**: 解析能力完整。

---

### bohrium-web-search (网页搜索) — v2 免费

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/search/web?q=&num=` | GET | ✅ | 返回 organic_results（冒烟实测 PASS） |

**计费**: v2 去掉了 v1 的 coding-plan 限额，免费。**结论**: 正常。

---

### bohrium-scholar-search (学者)

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/paper-server/scholar/search` (按姓名) | POST | ✅ | 正常（冒烟实测 PASS） |
| `/v2/paper-server/scholar/info?scholarId=` | GET | ✅ | 完整画像 |

**结论**: 按姓名搜索与详情查询正常。

---

### bohrium-sciencepedia (百科，原 bohrium-wiki) — 2026-06-26 真实实测（重写后 12 端点全通）

> 网关前缀 `/openapi/v2/literature-sage/wiki_v2/*` → LiteratureSage `/api/v1/wiki_v2`。
> 重写后 skill 围绕 4 个用户任务（搜词条/关键词、浏览领域课程、看课程章节、查知识图谱）。下表为 2026-06-26 用真实数据按调用链跑的冒烟，全部 HTTP=200 / `code=0`。其中 8 个核心免费端点（`/info`、`/search/universal`、`/article`、`/major_levels`、`/level_fields`、`/get_wiki_index`、`/knowledge_graph`、`/search_index_name`）已纳入仓库内 `tests/smoke_test.py` 的 `sciencepedia_smoke()` 链路（用前一步真实响应里的 id 串起来，不依赖硬编码）；图谱细节（`/knowledge_graph/node`、`/relationship`、`/search`）及 `/keyword` 暂以本表实测覆盖。

| 端点 | 方法 | 状态 | 实测要点（`data` 下） |
|------|------|------|------|
| `/wiki_v2/search/universal` | POST | ✅ | `articles[]`（`type`=article/keyword、`id`、`article_name`、`matched_elements`）+ `fields[]`（带 `field_id`/`node_id`，命中需用课程名查询） |
| `/wiki_v2/article` | POST | ✅ | `document` 全文：`article_name`/`main_content`/`seo_description`/`field_node.field_id` 等（用 `entry_id` 取，如 `many_body_physics-physics_of_graphene`） |
| `/wiki_v2/keyword` | POST | ✅ | `document`：`definition`/`applications`/`appendices`/`citations`/`current_revision_id` 等（用 `keyword_id` 取） |
| `/wiki_v2/major_levels` | POST | ✅ | `majors[]`（9 大类 / 17 分级），每级带 `node_id` |
| `/wiki_v2/level_fields` | POST | ✅ | 分页 `items[]`（`total`），`field` 对象含 `field_id`/`node_id`/`name`/`seo_title`、`topic_count`、示例 `topics` |
| `/wiki_v2/get_wiki_index` | POST | ✅ | 课程章节树 `wiki_indices` + `entry_count`/基础·核心·进阶计数；叶子 `entry` 带 `entry_id`/`snapshot`/`seo_description` |
| `/wiki_v2/knowledge_graph` | GET | ✅ | 从中心 `id` 展开：`nodes`(577)/`relationships`(999)/`domains`(8)，节点带 `node_type`/`field_id`/`depth` 等 |
| `/wiki_v2/knowledge_graph/search` | POST | ✅ | `items[]`（`type`=entry/keyword、`id`）用于拿图谱中心点 |
| `/wiki_v2/knowledge_graph/node` | GET | ✅ | 单节点详情（`display_name`/`description`/`field_name` 等） |
| `/wiki_v2/knowledge_graph/relationship` | GET | ✅ | 边详情，含 `evidences`（小写）/`evidence_count`/`is_cross_domain` |
| `/wiki_v2/info` | GET | ✅ | 总量：entry 47467 / keyword 108719 / total 156186 |
| `/wiki_v2/search_index_name` | POST | ✅ | 按名搜索索引（`node_types` 过滤；非课程名词可能 0 命中属正常） |

**结论**: 重写后 4 个用户任务覆盖的 12 个端点全部实测通过。`article`/`keyword` 用正确的 `entry_id`/`keyword_id` 可正常取全文（早先 250002 是该 id 无内容的个例，非接口问题）。修正点：`level_fields` 的 `field` 对象**确实返回 `field_id`**（Apifox 示例漏录），课程链接可直接用它拼。知识图谱为 **GET + 单个 `id`**（非 Apifox 标的 POST+ids）。

---

### bohrium-tools (科学工具库) — 2026-06-11 真实实测（9 端点全通）

> 网关前缀 `/openapi/v2/literature-sage/tool/*` → LiteratureSage 工具库（与 wiki 同上游）。实测 v1/v2 均返回 HTTP=200 / `code=0`，仓库统一走 v2。
> **响应包裹层（已实测确认）**：所有端点返回 `{"code": 0, "data": {...}, "trace_id": "..."}`，真实数据在 `data` 下，SKILL.md 示例已统一通过 `data()` 辅助函数解包。

| 端点 | 方法 | 状态 | 实测返回（`data` 下的键） |
|------|------|------|------|
| `/v2/literature-sage/tool/domain` | GET | ✅ | `items[]`：`node_id`/`node_name`/`tool_num`（如 Scientific AI Methods, 6274 个工具） |
| `/v2/literature-sage/tool/domain/summary` | GET | ✅ | `total_num` |
| `/v2/literature-sage/tool/subdomain` | POST | ✅ | `items` / `page` / `pageSize` / `total` |
| `/v2/literature-sage/tool/subdomain/detail` | POST | ✅ | `node_id` / `node_name` |
| `/v2/literature-sage/tool/list` | POST | ✅ | `items` / `page` / `pageSize` / `total`（如 DeepSpeed ★41838） |
| `/v2/literature-sage/tool/tags` | POST | ✅ | `items` |
| `/v2/literature-sage/tool/detail` | GET | ✅ | `name`/`star_count`/`fork_count`/`overview`/`tutorial`/`mcp_url`/`docker_image_uri`/`repo_url`/… |
| `/v2/literature-sage/tool/search/hybrid` | POST | ✅ | `tools` / `total`（注：实测 `tools[].score` 可能为 `null`，示例已做空值防护） |
| `/v2/literature-sage/tool/search/subdomain` | POST | ✅ | `subdomains` / `total` |

**结论**: 9 个端点（domain → subdomain → list/tags → detail，以及 hybrid / subdomain 检索）全部 HTTP=200 / `code=0`，`{code, data, trace_id}` 包裹层与文档一致；统一使用 v2 并解包 `data`。

---

### bohrium-lkm (大知识模型)

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v1/lkm/search` | POST | ✅ | 公开检索：claim/question/abstract 命中，支持 conclusion/premise scope、title/DOI/paper/date 过滤，默认按 paper 聚合 |
| `/v1/lkm/reasoning/search` | POST | ✅ | 推理链检索：支持 paper_ids/dois/title/date 过滤，推荐 `format=graph` |
| `/v1/lkm/claims/{id}/reasoning` | GET | ✅ | 单 claim 推理链，替代旧 `/evidence` 路径 |
| `/v1/lkm/papers/graph` | POST | ✅ | 论文级知识图谱，按 package_id/paper_id/doi/title 查询 |
| `/v1/lkm/variables/batch` | POST | ✅ | 批量水合节点详情 |
| `/v1/lkm/feedback` | POST | ✅ | 反馈提交；写入接口，不返回知识内容 |

**结论**: LKM 当前 skill 以 `/openapi/v1/lkm` 为准；旧 `/v2/lkm/claims/match`、`/claims/{id}/evidence`、`/papers/ocr/batch` 不是当前 skill 推荐路径。

---

### bohrium-mentor (Sigma 深度搜索) — 余额计费

> 网关 `/openapi/v2/sigma-search/*` → SigmaSearchProxy，附带 `SigmaBalanceBillMiddleware`（余额扣费，与 v1 的 SigmaBill 不同）。

| 端点 | 方法 | 状态 | 备注 |
|------|------|------|------|
| `/v2/sigma-search/api/v4/ai_search/sessions` | POST | ✅ | 创建会话成功，返回 sessionId（冒烟实测 PASS） |
| `/v2/sigma-search/api/v4/ai_search/sessions/{id}` | GET | ✅ | 会话详情（断线恢复用）；不存在的 id 返回 not found，证明路由可达 |
| `/v2/sigma-search/api/v3/sse/ai_search/v1/{id}/stream` | GET | ✅ | SSE 流式推理 |

**计费**: 创建会话按余额扣费。**结论**: 路由可用；冒烟测试会真实创建 session 并校验详情接口。

---

### bohrium-sandbox (沙箱 — lbg sdbx CLI)

| 功能 | 状态 | 备注 |
|------|------|------|
| `python sdbx.py doctor --json` | ✅ | sandbox_ready=true；只需 BOHR_ACCESS_KEY |
| `python sdbx.py template ls` / `list` | ✅ | 模板/沙箱列表正常 |
| `python sdbx.py create ... --json` | ✅ | 成功创建（计费） |
| `python sdbx.py exec` / `kill` | ✅ | 命令执行/销毁正常 |

**前置条件**: Python >=3.10；`python3 -m pip install --pre lbg`（4.0.0b*）。`sdbx.py` 将 `BOHR_ACCESS_KEY` 映射成 `BOHRIUM_ACCESS_KEY`。不经 `/openapi/vN` 网关，走 `launching/v2`。

**结论**: 全部功能正常；冒烟测试会真实 `create` 短时沙箱、`exec` 一条轻量命令，并在结束时 `kill --force` 清理。

---

## 文档与实际不符的问题汇总

| # | Skill | 问题 | 影响 | 状态 |
|---|-------|------|------|------|
| 1 | bohrium-job | API `POST /job/submit`、`GET /job_group/list` 返回 404 | 不影响 | CLI 正常，文档已说明优先 CLI |
| 2 | bohrium-node | API `/node/start/{id}` 返回 404 | 低 | 文档使用 `restart`；start 网关无路由 |
| 3 | ~~bohrium-dataset~~ | ~~`open.bohrium.com` 的 `POST /ds/` 307→404~~ | ~~中~~ | **已修复**：307 已修复，统一 open.bohrium.com |
| 4 | bohrium-wiki | article 端点返回 250002 "Article not found" | 低 | 索引存在，文章内容按需生成 |
| 5 | ~~bohrium-lkm~~ | ~~旧 `/papers/ocr/batch` 权限问题~~ | ~~低~~ | **已更新**：当前 skill 改用 `/openapi/v1/lkm` 的 search / reasoning / graph / feedback 路径，不再推荐 OCR 批处理 |

## CLI 环境配置说明

```bash
# 安装 Go 版 bohr CLI（job/node/job_group/dataset/image/project）
# macOS:
/bin/bash -c "$(curl -fsSL https://dp-public.oss-cn-beijing.aliyuncs.com/bohrctl/1.0.0/install_bohr_mac_curl.sh)"
# Linux:
/bin/bash -c "$(curl -fsSL https://dp-public.oss-cn-beijing.aliyuncs.com/bohrctl/1.0.0/install_bohr_linux_curl.sh)"

export PATH="$HOME/.bohrium:$PATH"
export OPENAPI_HOST=https://open.bohrium.com
export TIEFBLUE_HOST=https://tiefblue.dp.tech
export BOHR_ACCESS_KEY=<your_access_key>

# 安装 Python lbg CLI（sandbox，需 Python >=3.10）
python3 -m pip install --pre lbg   # 必须安装到 4.0.0b* prerelease
export BOHR_ACCESS_KEY=<your_access_key>
```

### bohr CLI 可用命令一览

| 模块 | 子命令 | 说明 |
|------|--------|------|
| `bohr job` | list, describe, log, download, submit, terminate, kill, delete | 任务全生命周期 |
| `bohr job_group` | list, create, terminate, delete, download | 任务组管理 |
| `bohr node` | list, create, stop, delete, connect | 开发机管理 |
| `bohr dataset` | list, create, delete | 数据集管理 |
| `bohr image` | list, pull, delete | 镜像管理 |
| `bohr project` | list | 项目列表 |
