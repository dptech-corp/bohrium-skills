#!/usr/bin/env python3
"""owners_table — 从 owners.yaml 渲染「认领表」，可直接粘进群里让各团队认领。

用法:
    python3 tools/owners_table.py            # markdown 表格
    python3 tools/owners_table.py --todo      # 只列尚未认领的模块
"""
import sys, os, yaml

REG = os.path.join(os.path.dirname(__file__), "..", "docs", "api", "owners.yaml")


def main():
    todo_only = "--todo" in sys.argv[1:]
    with open(REG, encoding="utf-8") as f:
        modules = yaml.safe_load(f)["modules"]

    rows = []
    for mod, info in modules.items():
        claimed = bool(info.get("team") or info.get("owner"))
        if todo_only and claimed:
            continue
        rows.append((
            mod,
            " ".join(info.get("path_prefixes", [])),
            info.get("team") or "❓待认领",
            info.get("owner") or "❓",
            info.get("contact") or "",
            info.get("service_repo") or "",
        ))

    print("| 模块 | 路径前缀 | 团队 | 负责人 | 联系方式(飞书/邮箱) | 后端仓库 |")
    print("|------|----------|------|--------|----------------------|----------|")
    for r in rows:
        print("| " + " | ".join(r) + " |")

    unclaimed = sum(1 for _, i in modules.items() if not (i.get("team") or i.get("owner")))
    print(f"\n> 共 {len(modules)} 个模块，{unclaimed} 个待认领。认领方式：改 "
          "`docs/api/owners.yaml` 里对应模块的 team/owner/contact/service_repo 并提 PR，"
          "或在群里回复由维护者代填。")


if __name__ == "__main__":
    main()
