"""Check complete public Skill packages against a published bohr artifact."""

import argparse
import json
import os
from pathlib import Path, PurePosixPath
import re
import shlex
import subprocess
import tempfile
from urllib.parse import unquote, urlsplit

FENCES = re.compile(r"```[^\n]*\n.*?```", re.S)
LINKS = re.compile(r"\]\(([^)]+)\)")
HEADINGS = re.compile(r"^#{1,6} +(.+)$", re.M)
NAME = re.compile(r"[a-z0-9]+(?:-[a-z0-9]+)*")


def require(condition, message):
    if not condition:
        raise AssertionError(message)


def package_files(root):
    return sorted(str(p.relative_to(root).as_posix()) for p in root.rglob("*") if p.is_file())


def adapt_source(text, entry, file, lock, proxy_host=""):
    if proxy_host:
        text = text.replace(proxy_host, lock["public_proxy_host"])
    if file == "SKILL.md":
        old = "name: " + entry["guide"] + "\n"
        require(text.count(old) == 1, f"{entry['guide']}: source frontmatter name changed")
        new = "name: " + entry["name"] + "\ncompatibility: " + json.dumps(lock["compatibility"]) + "\n"
        text = text.replace(old, new, 1)
    if file == "agents/openai.yaml":
        text = text.replace("$" + entry["guide"], "$" + entry["name"])
    for patch in lock.get("patches", []):
        if patch["guide"] == entry["guide"] and patch["path"] == file:
            require(text.count(patch["old"]) == 1, f"{entry['guide']}/{file}: source correction no longer matches")
            text = text.replace(patch["old"], patch["new"], 1)
    if file == "references/sandbox/network.md":
        text = lock["proxy_notice"] + text
    return text


def anchors(text):
    result = set(re.findall(r'<a id="([^"]+)"', text))
    counts = {}
    for heading in HEADINGS.findall(FENCES.sub("", text)):
        slug = re.sub(r"[^\w\- ]", "", heading.lower()).replace(" ", "-")
        count = counts.get(slug, 0)
        counts[slug] = count + 1
        result.add(slug if not count else f"{slug}-{count}")
    return result


def check_links(root):
    for file in root.rglob("*.md"):
        for target in LINKS.findall(FENCES.sub("", file.read_text())):
            if any(c in target for c in "{}<>"):
                continue
            target = target.split()[0]
            url = urlsplit(target)
            if url.scheme or url.netloc:
                continue
            destination = (file.parent / unquote(url.path)).resolve() if url.path else file.resolve()
            require(destination.is_relative_to(root.resolve()), f"{file}: link escapes package: {target}")
            require(destination.is_file(), f"{file}: missing local link: {target}")
            if url.fragment:
                require(unquote(url.fragment) in anchors(destination.read_text()), f"{file}: missing anchor: {target}")


def agent_fields(text):
    result = {}
    for key, value in re.findall(r"^  ([a-z_]+): (.+)$", text, re.M):
        result[key] = json.loads(value)
    require(set(result) == {"display_name", "short_description", "default_prompt"}, "incomplete agent interface metadata")
    require(all(isinstance(v, str) and v for v in result.values()), "empty agent interface field")
    return result


class CLI:
    def __init__(self, binary, env, repo):
        self.binary, self.env, self.repo = binary, env, repo

    def text(self, *args):
        result = subprocess.run([self.binary, *args], cwd=self.repo, env=self.env, text=True, capture_output=True, timeout=30)
        require(result.returncode == 0, f"bohr {' '.join(args)} exited {result.returncode}: {result.stderr}")
        return result.stdout

    def data(self, *args):
        result = json.loads(self.text(*args))
        require(result.get("ok") is True, f"bohr {' '.join(args)} did not return a success envelope")
        return result["data"]

    def tree(self, entry, version):
        guide = entry["guide"]
        pending, directories, files = [""], set(), {}
        while pending:
            directory = pending.pop()
            require(directory not in directories, f"{guide}: repeated source directory")
            directories.add(directory)
            location = guide + ("/" + directory if directory else "")
            listing = self.data("guide", "list", location, "-o", "json")
            require(listing.get("path") == location, f"{guide}: incorrect source listing")
            for item in listing["items"]:
                prefix = guide + "/"
                require(item["path"].startswith(prefix), f"{guide}: source path outside guide")
                path = item["path"][len(prefix):]
                require(path and ".." not in PurePosixPath(path).parts and not path.startswith("/"), f"{guide}: invalid source path")
                if item["is_dir"]:
                    pending.append(path)
                    continue
                require(path not in files, f"{guide}: duplicate source file")
                data = self.data("guide", "read", guide, path, "-o", "json")
                require((data["name"], data["version"], data["path"]) == (guide, version, path), f"{guide}/{path}: source identity/version mismatch")
                files[path] = data["content"]
            require(len(files) + len(directories) < 10000, "source tree exceeds the supported package size")
        require("SKILL.md" in files and any(p.startswith("references/") for p in files), f"{guide}: missing native main/reference files")
        explicit = self.data("guide", "read", guide, "-o", "json")
        default = self.data("guide", "read", guide)
        require(explicit == default, f"{guide}: default non-interactive output differs from JSON")
        require(explicit["content"] == files["SKILL.md"], f"{guide}: default guide file differs")
        return files


def check_workflow_previews(repo, cli):
    # Exercise the documented config-to-request boundary with the published
    # binary, without submitting a job or uploading any input.
    body = (repo / "en/bohrium-job/SKILL.md").read_text()
    config = json.loads(re.search(r"```json\n(.*?)\n```", body, re.S).group(1))
    with tempfile.TemporaryDirectory(prefix="bohr-workflow-preview-") as directory:
        fixture = Path(directory)
        config_path = fixture / "job.json"
        config_path.write_text(json.dumps(config))
        inputs = fixture / "input"
        inputs.mkdir()
        (inputs / "run.py").write_text("print('offline preview input')\n")
        preview = cli.data(
            "job", "submit", "-i", str(config_path), "--project-id", "1234",
            "--job_name", "analysis", "--image_address", "registry.example.com/research:v1",
            "--machine_type", "preview-machine", "--command", "python run.py > run.log 2>&1",
            "--input_directory", str(inputs), "--result_path", "/personal/results",
            "--dry-run", "-o", "json",
        )
        require(preview.get("dry_run") is True, "job recipe unexpectedly executed")
        operations = preview["operations"]
        request = next(item["body"] for item in operations if item["path"] == "/openapi/v4/job/add")
        require(request["outFiles"] == ["result.csv"] and request["logFiles"] == ["run.log"], "job recipe dropped its output/log collection contract")
        require(request["downloadPath"] == "/personal/results/" and request["projectId"] == 1234, "job recipe lost remote result/project scope")
        require(any(item["method"] == "UPLOAD" and item["path"] == str(inputs) for item in operations), "job recipe did not preserve its explicit input tree")

    # Take the actual argv example from the guide so putting CLI flags after
    # the remote delimiter becomes an observable regression.
    sandbox = (repo / "en/bohrium-sandbox/SKILL.md").read_text()
    line = next(line for line in sandbox.splitlines() if line.startswith("bohr sandbox exec <sandbox_id> -o json -- "))
    args = shlex.split(line.replace("<sandbox_id>", "default--preview-abc12"))[1:]
    delimiter = args.index("--")
    args.insert(delimiter, "--dry-run")
    preview = cli.data(*args)
    require(preview.get("dry_run") is True, "sandbox recipe unexpectedly executed")
    request = preview["operations"][0]["body"]
    require("/tmp/run.py" in request["command"] and "json" not in request["command"] and "-o" not in request["command"], "sandbox recipe passed local output flags to the remote script")
    require(request["background"] is False, "sandbox recipe unexpectedly launches background work")
    print("Verified documented Job output collection and Sandbox argv handoffs with offline published-CLI previews.")


def check_packages(repo, lock, cli):
    version = cli.text("--version").strip()
    require(version == "bohr version " + lock["version"], f"wrong CLI version: {version}")
    require(lock["package"] == "@dptech-corp/bohr-cli", "unexpected source package")
    require(lock["skills"], "empty source selection")
    expected_names = {entry["name"] for entry in lock["skills"]}
    require(len(expected_names) == len(lock["skills"]), "duplicate source selection")
    for lang in ("en", "zh"):
        names = {p.name for p in (repo / lang).iterdir() if p.is_dir() and p.name.startswith("bohrium-")}
        require(names == expected_names, f"{lang}: public Skill names differ from source selection")
    total = 0
    for entry in lock["skills"]:
        name, guide = entry["name"], entry["guide"]
        require(NAME.fullmatch(name) and len(name) <= 64 and NAME.fullmatch(guide), "invalid Skill identity")
        source = cli.tree(entry, lock["version"])
        require(sorted(source) == entry["files"], f"{name}: published source inventory changed")
        proxy_host = ""
        network = source.get("references/sandbox/network.md")
        if network:
            match = re.search(r"http://([^/\s:]+):3128", network)
            require(match is not None, "source proxy documentation changed")
            proxy_host = match.group(1)
        for lang in ("en", "zh"):
            root = repo / lang / name
            require(package_files(root) == sorted(source), f"{lang}/{name}: missing or extra package files")
            for file, content in source.items():
                expected = adapt_source(content, entry, file, lock, proxy_host)
                actual = (root / file).read_text()
                if lang == "en" or file not in ("SKILL.md", "agents/openai.yaml"):
                    require(actual == expected, f"{lang}/{name}/{file}: differs from native source/adaptations")
                if file == "SKILL.md":
                    require(re.search(r"^name: " + re.escape(name) + r"$", actual, re.M), f"{lang}/{name}: frontmatter identity mismatch")
                if file == "agents/openai.yaml":
                    fields = agent_fields(actual)
                    require("$" + name in fields["default_prompt"], f"{lang}/{name}: agent prompt uses the wrong Skill")
                if proxy_host:
                    require(proxy_host not in actual, f"{lang}/{name}: environment-specific proxy host leaked")
            check_links(root)
        en = (repo / "en" / name / "SKILL.md").read_text()
        zh = (repo / "zh" / name / "SKILL.md").read_text()
        require(re.search(r"[\u4e00-\u9fff]", zh), f"{name}: Chinese main document was not localized")
        require(FENCES.findall(en) == FENCES.findall(zh), f"{name}: Chinese command examples differ from native English")
        require(len(HEADINGS.findall(en)) == len(HEADINGS.findall(zh)), f"{name}: Chinese workflow sections are missing")
        cli.text(*entry["command"].split()[1:], "--help")
        total += len(source)
    guardrails = {
        "bohrium-paper-search": ["--page", "--size", "BILLING_INSUFFICIENT", "--payment-source photon", "--payment-source balance", "0.05", "0.50"],
        "bohrium-pdf-parser": ["data.token", "undefined", "processing", "success", "PARSE_FAILED", "--return-half", "proc_page", "total_page"],
        "bohrium-sandbox": ["--create-request-id", "sandbox_create_outcome_uncertain", "--wenyon-mode", "--commit", "--discard", "produced_version", "data.ready", "capabilities.dynamic_adjustment.available=true", 'REQUEST_ID="$(uuidgen)"', '--request-id "$REQUEST_ID"'],
    }
    for lang in ("en", "zh"):
        for name, tokens in guardrails.items():
            body = (repo / lang / name / "SKILL.md").read_text()
            require(all(token in body for token in tokens), f"{lang}/{name}: native workflow boundary missing")
    for name in ("bohrium-paper-search", "bohrium-sandbox"):
        with tempfile.TemporaryDirectory(prefix="bohr-native-install-") as project:
            receipt = cli.data("skill", "install", str(repo / "en" / name), "--target", "codex", "--project-dir", project, "-o", "json")
            installed = Path(project) / ".agents" / "skills" / name
            original = repo / "en" / name
            require(receipt.get("path") == str(installed), f"{name}: installer receipt destination mismatch")
            require(package_files(installed) == package_files(original), f"{name}: published installer dropped resources")
            for file in package_files(original):
                require((installed / file).read_bytes() == (original / file).read_bytes(), f"{name}/{file}: installed bytes changed")
    for command in ("search", "patent"):
        previews = [cli.data("paper", command, "pagination contract check", "--page", page, "--size", "3", "--type", "0", "--dry-run", "-o", "json") for page in ("1", "2")]
        require(all(item.get("dry_run") is True and len(item["operations"]) == 1 for item in previews), "paper dry-run contract changed")
        first, second = [item["operations"][0] for item in previews]
        require(first == second and first["body"]["pageSize"] == 3, f"{command}: page forwarding changed; review the pinned limitation/adaptation")
        require(not first.get("params") and "page" not in first["body"], f"{command}: published page limitation changed")
    print("Confirmed pinned paper/patent page 1 and 2 produce identical request previews without a page parameter; no gateway call was made.")
    help_text = cli.text("sandbox", "create", "--help")
    require(all(flag in help_text for flag in ("--wenyon-mode", "--wenyon-version", "--wenyon-user-token-env")), "published CLI lacks documented instance-mount flags")
    delete_help = cli.text("sandbox", "delete", "--help")
    require("--commit" in delete_help and "--discard" in delete_help, "published CLI lacks documented settlement flags")
    check_workflow_previews(repo, cli)
    print(f"Verified {len(lock['skills'])} native Skills, {total} source files per language, local links, localized examples, metadata, command help, and complete isolated installs.")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--cli", required=True)
    args = parser.parse_args()
    repo = Path(__file__).resolve().parents[1]
    lock = json.loads((repo / "cli-skills.lock.json").read_text())
    with tempfile.TemporaryDirectory(prefix="bohr-native-config-") as config:
        env = os.environ.copy()
        env.update({"HOME": config, "BOHR_CONFIG_DIR": config, "BOHR_ACCESS_KEY": ""})
        check_packages(repo, lock, CLI(str(Path(args.cli).absolute()), env, repo))


if __name__ == "__main__":
    main()
