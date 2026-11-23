import csv
from pathlib import Path
import subprocess
import tempfile
import unittest

from container_workflow import run_workflow


class FakeDocker:
    def __init__(self, fail=None):
        self.calls = []
        self.fail = fail

    def __call__(self, *args, check=True):
        self.calls.append(args)
        if args[0] == "build":
            iid = Path(args[args.index("--iidfile") + 1])
            iid.write_text("sha256:test-image\n", encoding="ascii")
        if self.fail in args:
            raise subprocess.CalledProcessError(17, args)
        return subprocess.CompletedProcess(args, 0)


def mount_fields(call):
    return next(csv.reader([call[call.index("--mount") + 1]]))


class ContainerWorkflowTests(unittest.TestCase):
    def test_seeded_lifecycle(self):
        docker = FakeDocker()
        run_workflow(invoke=docker)
        build, create, seed, test, remove = docker.calls
        self.assertEqual(build[0], "build")
        self.assertEqual(create[:2], ("volume", "create"))
        volume = create[-1]
        self.assertTrue(volume.startswith("git-view-fixture-"))
        for call in (seed, test):
            self.assertIn("--network=none", call)
            self.assertIn("sha256:test-image", call)
            self.assertIn("type=volume", mount_fields(call))
            self.assertIn("source=" + volume, mount_fields(call))
        self.assertNotIn("readonly", mount_fields(seed))
        self.assertIn("readonly", mount_fields(test))
        self.assertEqual(remove, ("volume", "rm", volume))

    def test_failures_remove_the_seeded_volume(self):
        for script in ("/src/tests/container/seed.sh", "/src/tests/container/check.sh"):
            with self.subTest(script=script):
                docker = FakeDocker(fail=script)
                with self.assertRaises(subprocess.CalledProcessError) as failure:
                    run_workflow(invoke=docker)
                self.assertEqual(failure.exception.returncode, 17)
                self.assertEqual(docker.calls[-1], ("volume", "rm", docker.calls[1][-1]))

    def test_host_checkout_is_bound_read_only_without_a_seed_volume(self):
        docker = FakeDocker()
        with tempfile.TemporaryDirectory(prefix="repo with spaces, comma ") as directory:
            repo = Path(directory).resolve()
            run_workflow(repo=repo, invoke=docker)
        build, test = docker.calls
        self.assertEqual(build[0], "build")
        self.assertIn("--network=none", test)
        self.assertIn("sha256:test-image", test)
        self.assertEqual(
            mount_fields(test),
            ["type=bind", "source=" + str(repo), "target=/repo", "readonly", "bind-recursive=disabled"],
        )
        self.assertEqual(test[-1], "host")

    def test_missing_host_directory_fails_before_build(self):
        docker = FakeDocker()
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(ValueError):
                run_workflow(repo=Path(directory) / "missing", invoke=docker)
        self.assertEqual(docker.calls, [])


if __name__ == "__main__":
    unittest.main()
