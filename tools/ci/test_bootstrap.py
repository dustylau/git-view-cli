import hashlib
import io
import unittest
import zipfile

from bootstrap import candidate_run, verify_artifact


class BootstrapTests(unittest.TestCase):
    def test_verifies_archives_from_the_selected_revision(self):
        for version in ("0.1.0", "1.0.0"):
            with self.subTest(version=version):
                buffer = io.BytesIO()
                with zipfile.ZipFile(buffer, "w") as archive:
                    manifest = []
                    for target in ("linux-amd64", "linux-arm64", "windows-amd64", "windows-arm64"):
                        name = f"git-view-{version}-ci.abc-{target}.zip"
                        data = target.encode("ascii")
                        archive.writestr(name, data)
                        manifest.append(f"{hashlib.sha256(data).hexdigest()}  {name}\n")
                    archive.writestr("SHA256SUMS", "".join(manifest))
                verify_artifact(buffer.getvalue(), "abc")
                with self.assertRaises(ValueError):
                    verify_artifact(buffer.getvalue(), "another-commit")

    def test_selects_only_the_local_pipeline_for_this_push(self):
        match = {
            "head_sha": "commit", "head_branch": "dev/topic", "event": "push",
            "path": "local-ci.yml@refs/heads/dev/topic",
        }
        runs = [
            dict(match, head_sha="old"),
            dict(match, head_branch="other"),
            dict(match, event="pull_request"),
            dict(match, path="other.yml@refs/heads/dev/topic"),
            match,
        ]
        self.assertIs(candidate_run(runs, "commit", "dev/topic"), match)
        self.assertIsNone(candidate_run(runs[:-1], "commit", "dev/topic"))
