import tempfile
from pathlib import Path
import unittest

import check_native_skills as check


class NativeDistributionChecks(unittest.TestCase):
    def test_example_validation_forces_help_and_excludes_remote_argv(self):
        text = '```bash\nbohr sandbox exec <sandbox_id> -o json -- python run.py --not-a-cli-flag\n```\n'
        self.assertEqual(list(check.bash_help_arguments(text)), [("sandbox", "exec", "1", "-o", "json", "--help")])

    def test_example_cannot_disable_forced_help(self):
        text = '```bash\nbohr image list --help=false -h=false --page <page>\n```\n'
        self.assertEqual(list(check.bash_help_arguments(text)), [("image", "list", "--page", "1", "--help")])

    def test_continuations_variables_quotes_and_redirection_are_not_executed(self):
        text = '```bash\nbohr job submit --project-id "$PROJECT_ID" \\\n --command "python run.py > run.log" > receipt.json\n```\n'
        self.assertEqual(list(check.bash_help_arguments(text)), [("job", "submit", "--project-id", "1", "--command", "python run.py > run.log", "--help")])

    def test_dynamic_routing_and_non_shell_fences_are_not_executable_examples(self):
        text = '```bash\nbohr sandbox <subcommand> --help\n```\n```json\nbohr ignored\n```\n'
        self.assertEqual(list(check.bash_help_arguments(text)), [])

    def test_multiline_remote_shell_text_stays_one_help_argument(self):
        text = '```bash\nbohr sandbox exec <id> -c "echo first\necho second"\n```\n'
        self.assertEqual(list(check.bash_help_arguments(text)), [("sandbox", "exec", "1", "-c", "echo first\necho second", "--help")])

    def test_adaptation_preserves_cli_commands_and_changes_only_public_identity(self):
        source = '---\nname: bohr-demo\ndescription: Demo\n---\n\n# Demo\n\nbohr guide read bohr-demo\n'
        entry = {"name": "bohrium-demo", "guide": "bohr-demo"}
        lock = {"compatibility": "Requires bohr", "patches": []}
        actual = check.adapt_source(source, entry, "SKILL.md", lock)
        self.assertIn('name: bohrium-demo\ncompatibility: "Requires bohr"\n', actual)
        self.assertIn("bohr guide read bohr-demo\n", actual)
        prompt = 'interface:\n  default_prompt: "Use $bohr-demo"\n'
        self.assertIn("$bohrium-demo", check.adapt_source(prompt, entry, "agents/openai.yaml", lock))

    def test_changed_upstream_patch_is_rejected(self):
        entry = {"name": "bohrium-demo", "guide": "bohr-demo"}
        lock = {"patches": [{"guide": "bohr-demo", "path": "references/commands.md", "old": "old contract", "new": "new contract"}]}
        with self.assertRaisesRegex(AssertionError, "no longer matches"):
            check.adapt_source("different upstream contract", entry, "references/commands.md", lock)

    def test_missing_reference_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "SKILL.md").write_text("[details](references/commands.md)\n")
            with self.assertRaisesRegex(AssertionError, "missing local link"):
                check.check_links(root)

    def test_reference_cannot_escape_installed_package(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "SKILL.md").write_text("[details](../outside.md)\n")
            with self.assertRaisesRegex(AssertionError, "escapes package"):
                check.check_links(root)

    def test_localized_anchor_is_resolved_and_missing_anchor_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "SKILL.md").write_text('<a id="payment-source"></a>\n\n# 支付来源\n')
            ref = root / "reference.md"
            ref.write_text("[payment](SKILL.md#payment-source)\n")
            check.check_links(root)
            ref.write_text("[payment](SKILL.md#missing-anchor)\n")
            with self.assertRaisesRegex(AssertionError, "missing anchor"):
                check.check_links(root)


if __name__ == "__main__":
    unittest.main()
