"""Exercise real consumer discovery without credentials or model requests."""

import argparse
from collections import Counter
import json
import os
from pathlib import Path
import queue
import subprocess
import tempfile
import threading
import time


def run(args, env, cwd, timeout=120):
    result = subprocess.run(args, env=env, cwd=cwd, capture_output=True, text=True, timeout=timeout)
    if result.returncode:
        raise AssertionError(f"{args[0]} {' '.join(args[1:])} exited {result.returncode}: {result.stdout}\n{result.stderr}")
    return result.stdout


def rpc_skills(binary, env, cwd):
    messages = queue.Queue()
    with tempfile.TemporaryFile(mode="w+") as errors:
        process = subprocess.Popen([binary, '-c', 'skills.bundled.enabled=false', 'app-server', '--listen', 'stdio://'], env=env, cwd=cwd, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=errors, text=True, bufsize=1)
        def read_messages():
            for line in process.stdout:
                try:
                    messages.put(json.loads(line))
                except ValueError:
                    messages.put({"protocolError": line})
            messages.put(None)
        thread = threading.Thread(target=read_messages, daemon=True)
        thread.start()
        def send(message):
            process.stdin.write(json.dumps(message) + '\n')
            process.stdin.flush()
        def receive(request_id):
            deadline = time.monotonic() + 45
            while True:
                message = messages.get(timeout=max(0.01, deadline-time.monotonic()))
                if message is None or 'protocolError' in message:
                    errors.seek(0)
                    raise AssertionError(f"Codex protocol stopped: {message}\n{errors.read()}")
                if message.get('id') == request_id:
                    if 'error' in message:
                        raise AssertionError(f"Codex request {request_id} rejected: {message['error']}")
                    return message['result']
        try:
            send({"jsonrpc":"2.0", "id":1, "method":"initialize", "params":{"clientInfo":{"name":"bohrium-skill-consumer-check", "version":"1.0.0"}, "capabilities":{"experimentalApi":True}}})
            receive(1)
            send({"jsonrpc":"2.0", "method":"initialized"})
            send({"jsonrpc":"2.0", "id":2, "method":"skills/list", "params":{"cwds":[str(cwd)], "forceReload":True}})
            return receive(2)
        finally:
            process.stdin.close()
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.terminate()
                process.wait(timeout=10)
            thread.join(timeout=1)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--installer', required=True)
    parser.add_argument('--tools-prefix', required=True)
    args = parser.parse_args()
    repo = Path(__file__).resolve().parents[1]
    tools = Path(args.tools_prefix).resolve()
    names = {item['name'] for item in json.loads((repo/'cli-skills.lock.json').read_text())['skills']}
    bins = tools/'node_modules'/'.bin'
    with tempfile.TemporaryDirectory(prefix='bohrium-consumers-') as temp:
        home = Path(temp)/'home'; home.mkdir()
        project = Path(temp)/'project'; project.mkdir(); (project/'.git').mkdir()
        env = os.environ.copy()
        env.update({'HOME':str(home), 'CODEX_HOME':str(home/'.codex'), 'CLAUDE_CONFIG_DIR':str(home/'.claude'), 'DSH_HOME':str(home/'.dsh'), 'DSH_AGENTS_HOME':str(home/'.agents'), 'CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC':'1', 'DISABLE_TELEMETRY':'1', 'DISABLE_ERROR_REPORTING':'1', 'CODEX_APP_SERVER_MANAGED_CONFIG_PATH':str(home/'managed_config.toml')})
        for key in ['OPENAI_API_KEY','ANTHROPIC_API_KEY','ANTHROPIC_AUTH_TOKEN','DEEPSEEK_API_KEY','DSH_BUNDLED_SKILL_DIR']:
            env.pop(key, None)
        # Exercise the real upgrade route, not only a clean canonical directory.
        legacy = home/'.codex/skills'
        for name in names:
            leaf = legacy/name
            leaf.mkdir(parents=True)
            (leaf/'SKILL.md').write_text(f'---\nname: {name}\ndescription: Legacy API Skill\n---\nLegacy content\n')
            (leaf/'user-note.txt').write_text('preserve this user addition')
        custom = legacy/'my-custom-skill'
        custom.mkdir()
        (custom/'SKILL.md').write_text('---\nname: my-custom-skill\ndescription: Unrelated user Skill\n---\nUser content\n')
        receipt = json.loads(run([str(Path(args.installer).resolve()), 'install', '--home', str(home), '--lang', 'en', '--json'], env, repo))
        assert set(receipt['official']) == names
        assert set(receipt['migrated_legacy']) == names
        assert all(not (legacy/name).exists() for name in names), 'Legacy duplicate Skills remain active'
        assert (custom/'SKILL.md').read_text().endswith('User content\n')
        notes = list((home/'.config/bohrium-skills-cli/backups').rglob('user-note.txt'))
        assert len(notes) == len(names) and all(p.read_text() == 'preserve this user addition' for p in notes)
        failures = []
        try:
            result = rpc_skills(str(bins/'codex'), env, project)
            assert len(result['data']) == 1
            entry = result['data'][0]
            assert not entry['errors'], entry['errors']
            selected = [skill for skill in entry['skills'] if skill['name'] in names]
            counts = Counter(skill['name'] for skill in selected)
            assert set(counts) == names, f"Codex missing public Skills: {names-set(counts)}"
            assert all(value == 1 for value in counts.values()), f"Codex duplicate public Skill names: {dict(counts)}"
            for skill in selected:
                path = Path(skill['path'])
                assert path.is_file() and path.name == 'SKILL.md'
                if skill['name'] != 'bohrium-sandbox':
                    interface = skill.get('interface')
                    assert interface and '$'+skill['name'] in interface['defaultPrompt'], f"Codex did not load interface identity for {skill['name']}"
                assert (path.parent/'references').is_dir()
            print(f"Codex real app-server discovered {len(selected)} complete Skills with interface metadata.")
            alias_home = Path(temp)/'alias-home'
            canonical = alias_home/'.agents/skills'
            canonical.mkdir(parents=True)
            (alias_home/'.codex').mkdir()
            (alias_home/'.codex/skills').symlink_to(canonical, target_is_directory=True)
            alias_env = env.copy()
            alias_env.update({'HOME':str(alias_home), 'CODEX_HOME':str(alias_home/'.codex'), 'CLAUDE_CONFIG_DIR':str(alias_home/'.claude'), 'DSH_HOME':str(alias_home/'.dsh'), 'DSH_AGENTS_HOME':str(alias_home/'.agents'), 'CODEX_APP_SERVER_MANAGED_CONFIG_PATH':str(alias_home/'managed_config.toml')})
            run([str(Path(args.installer).resolve()), 'install', '--home', str(alias_home), '--lang', 'en', '--json'], alias_env, repo)
            alias_result = rpc_skills(str(bins/'codex'), alias_env, project)
            alias_counts = Counter(skill['name'] for item in alias_result['data'] for skill in item['skills'] if skill['name'] in names)
            assert set(alias_counts) == names and all(value == 1 for value in alias_counts.values()), f'Codex canonical root alias duplicates/loss: {dict(alias_counts)}'
            print('Codex whole-root symlink aliases preserve one complete Skill per name.')

        except Exception as error:
            failures.append(f"Codex: {type(error).__name__}: {error}")
        try:
            print(run(['node', str(repo/'scripts/check_dsh_consumer.mjs'), str(tools), str(home), str(project), str(repo)], env, repo))
        except Exception as error:
            failures.append(f"DSH: {type(error).__name__}: {error}")
        try:
            claude = str(bins/'claude')
            print(run([claude, 'plugin', 'validate', str(repo)], env, repo))
            print(run([claude, 'plugin', 'marketplace', 'add', str(repo)], env, repo))
            print(run([claude, 'plugin', 'install', 'bohrium-skills@bohrium', '--scope', 'user'], env, repo))
            listing = json.loads(run([claude, 'plugin', 'list', '--json'], env, repo))
            installed = json.loads((home/'.claude/plugins/installed_plugins.json').read_text())
            records = installed['plugins']['bohrium-skills@bohrium']
            assert records, 'Claude did not record the plugin installation'
            for record in records:
                root = Path(record['installPath'])
                for name in names:
                    assert (root/'en'/name/'SKILL.md').is_file(), f"Claude cache missing {name}"
                    assert (root/'en'/name/'references').is_dir(), f"Claude cache missing references for {name}"
            assert 'bohrium-skills@bohrium' in json.dumps(listing), 'Claude plugin list does not show the installed plugin'
            print('Claude real plugin validator and installer accepted the marketplace and cached all 17 complete Skills.')
        except Exception as error:
            failures.append(f"Claude: {type(error).__name__}: {error}")
        if failures:
            raise AssertionError('\n'.join(failures))


if __name__ == '__main__':
    main()
