from pathlib import Path
import subprocess
import tempfile
import unittest

from package_release import package_release


class PackageReleaseTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.out = Path(self.scratch.name) / "dist"

    def test_invalid_version_does_not_create_output(self):
        with self.assertRaises(ValueError):
            package_release("../0.1.0", self.out, build=self.unexpected_build)
        self.assertFalse(self.out.exists())

    def test_existing_candidate_is_preserved(self):
        existing = self.out / "git-view-0.1.0-rc.1"
        existing.mkdir(parents=True)
        marker = existing / "existing.txt"
        marker.write_bytes(b"existing")
        with self.assertRaises(FileExistsError):
            package_release("0.1.0-rc.1", self.out, build=self.unexpected_build)
        self.assertEqual(marker.read_bytes(), b"existing")

    def test_failed_build_leaves_no_partial_candidate(self):
        def fail_second_build(os_name, arch, version, binary):
            if arch == "arm64":
                raise subprocess.CalledProcessError(1, ["go", "build"])
            binary.write_bytes(b"first binary")

        with self.assertRaises(subprocess.CalledProcessError):
            package_release("0.1.0-rc.1", self.out, build=fail_second_build)
        self.assertEqual(list(self.out.iterdir()), [])

    def unexpected_build(self, *args):
        self.fail("validation should have stopped before calling Go")


if __name__ == "__main__":
    unittest.main()
