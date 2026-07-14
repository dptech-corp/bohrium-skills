#!/usr/bin/env python3
"""whoowns — 从接口路径或一段日志定位负责模块与团队。

用法:
    python3 tools/whoowns.py "/openapi/v2/parse/get-result"
    python3 tools/whoowns.py "500 error at POST /openapi/v2/knowledge/file/submit ..."
    echo "<日志>" | python3 tools/whoowns.py            # 从 stdin 读
    python3 tools/whoowns.py --json "/openapi/v1/lkm/search"   # 机读输出

按 path_prefixes 做「最长前缀匹配」，未命中则报告 unknown。
注册表: docs/api/owners.yaml
"""
import sys, os, re, json

try:
    import yaml
except ImportError:
    sys.exit("需要 PyYAML：pip install pyyaml")

REG = os.path.join(os.path.dirname(__file__), "..", "docs", "api", "owners.yaml")
PATH_RE = re.compile(r"/openapi/[A-Za-z0-9_./{}-]+")


def load_modules():
    with open(REG, encoding="utf-8") as f:
        return yaml.safe_load(f)["modules"]


def extract_path(text):
    m = PATH_RE.search(text)
    return m.group(0) if m else text.strip()


def match(path, modules):
    best, best_len = None, -1
    for name, info in modules.items():
        for pref in info.get("path_prefixes", []):
            if path.startswith(pref) and len(pref) > best_len:
                best, best_len = (name, info), len(pref)
    return best


def main():
    args = [a for a in sys.argv[1:] if a != "--json"]
    as_json = "--json" in sys.argv[1:]
    raw = " ".join(args) if args else sys.stdin.read()
    if not raw.strip():
        sys.exit("用法: whoowns.py \"<路径或日志>\"")
    path = extract_path(raw)
    modules = load_modules()
    hit = match(path, modules)
    if not hit:
        out = {"path": path, "module": None, "message": "未匹配到任何模块，请检查路径或补充 owners.yaml"}
        print(json.dumps(out, ensure_ascii=False) if as_json else out["message"])
        sys.exit(2)
    name, info = hit
    result = {
        "path": path, "module": name, "tag": info.get("tag"),
        "team": info.get("team") or "(未认领)",
        "owner": info.get("owner") or "(未认领)",
        "contact": info.get("contact") or "",
        "service_repo": info.get("service_repo") or "",
        "oncall": info.get("oncall") or "",
        "skill_doc": info.get("skill_doc"),
    }
    if as_json:
        print(json.dumps(result, ensure_ascii=False, indent=2))
    else:
        print(f"路径:      {result['path']}")
        print(f"模块:      {result['module']}  [{result['tag']}]")
        print(f"团队/负责人: {result['team']} / {result['owner']}")
        if result["contact"]:
            print(f"联系方式:  {result['contact']}")
        if result["service_repo"]:
            print(f"后端仓库:  {result['service_repo']}")
        if result["oncall"]:
            print(f"值班:      {result['oncall']}")
        print(f"契约文档:  {result['skill_doc']}  (git blame 可查最近改动人)")


if __name__ == "__main__":
    main()
