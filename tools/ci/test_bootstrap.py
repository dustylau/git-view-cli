import unittest

from bootstrap import candidate_run


class BootstrapTests(unittest.TestCase):
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
