"""Seed a Gitea repository and optionally wait for local delivery."""

import argparse
import base64
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time
from urllib.error import HTTPError
from urllib.parse import urlencode, urlsplit
from urllib.request import Request, urlopen
import zipfile


SERVER = "http://git:3000"
OWNER = "local-ci"
WORKFLOW = ".gitea/workflows/local-ci.yml"
ARTIFACT = "git-view-candidate"


def git(*args, authorization=None):
    env = dict(os.environ, GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL="/dev/null", GIT_TERMINAL_PROMPT="0")
    if authorization:
        env.update(GIT_CONFIG_COUNT="1", GIT_CONFIG_KEY_0=f"http.{SERVER}/.extraHeader",
                   GIT_CONFIG_VALUE_0=f"Authorization: {authorization}")
    return subprocess.check_output(
        ["git", "-C", "/checkout", "-c", "safe.directory=/checkout", "-c", "core.hooksPath=/dev/null",
         "-c", "credential.helper=", *args], env=env, text=True, encoding="utf-8",
    ).strip()


def resolve_revision(ref, branch):
    sha = git("rev-parse", "--verify", "--end-of-options", f"{ref}^{{commit}}")
    if branch is None:
        symbolic = git("rev-parse", "--symbolic-full-name", ref)
        branch = symbolic.removeprefix("refs/heads/") if symbolic.startswith("refs/heads/") else "local-ci"
    git("check-ref-format", f"refs/heads/{branch}")
    return sha, branch


def api(path, authorization, method="GET", data=None):
    headers = {"Authorization": authorization, "Accept": "application/json"}
    if data is not None:
        headers["Content-Type"] = "application/json"
        data = json.dumps(data).encode("utf-8")
    request = Request(f"{SERVER}/api/v1{path}", data=data, headers=headers, method=method)
    with urlopen(request, timeout=15) as response:
        return response.read()


def browser_url(url):
    return os.environ["LOCAL_CI_URL"] + urlsplit(url).path


def ensure_repository(name, branch, authorization):
    path = f"/repos/{OWNER}/{name}"
    try:
        repo = json.loads(api(path, authorization))
    except HTTPError as err:
        if err.code != 404:
            raise
        repo = json.loads(api("/user/repos", authorization, "POST", {
            "name": name, "private": True, "auto_init": False, "default_branch": branch,
        }))
    if not repo["private"]:
        raise ValueError(f"Use a private local repository for {OWNER}/{name}.")
    if not repo["has_actions"]:
        api(path, authorization, "PATCH", {"has_actions": True})
    return browser_url(repo["html_url"])


def candidate_run(runs, sha, branch):
    workflow_ref = f"{Path(WORKFLOW).name}@refs/heads/{branch}"
    return next((run for run in runs if run["head_sha"] == sha and run["head_branch"] == branch
                 and run["event"] == "push" and run["path"] == workflow_ref), None)


def verify_artifact(content, sha):
    with zipfile.ZipFile(io.BytesIO(content)) as delivered:
        pattern = rf"(git-view-[0-9]+\.[0-9]+\.[0-9]+-ci\.{re.escape(sha)})-linux-amd64\.zip"
        stems = [m[1] for name in delivered.namelist() if (m := re.fullmatch(pattern, name))]
        if len(stems) != 1:
            raise ValueError("Expected one Linux amd64 ZIP for the selected commit.")
        stem = stems[0]
        archives = [f"{stem}-{target}.zip" for target in ("linux-amd64", "linux-arm64", "windows-amd64", "windows-arm64")]
        if sorted(delivered.namelist()) != sorted(["SHA256SUMS", *archives]):
            raise ValueError("Delivered artifact does not contain the expected ZIPs and manifest.")
        expected = "".join(f"{hashlib.sha256(delivered.read(name)).hexdigest()}  {name}\n" for name in archives)
        if delivered.read("SHA256SUMS").decode("ascii") != expected:
            raise ValueError("Delivered ZIP checksums do not match the manifest.")


def wait_for_delivery(name, sha, branch, authorization, timeout):
    repo_path = f"/repos/{OWNER}/{name}"
    query = urlencode({"branch": branch, "head_sha": sha, "event": "push"})
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        runs = json.loads(api(f"{repo_path}/actions/runs?{query}", authorization))["workflow_runs"]
        run = candidate_run(runs, sha, branch)
        if run and run["status"] == "completed":
            url = browser_url(run["html_url"])
            if run["conclusion"] != "success":
                raise ValueError(f"Actions run ended with {run['conclusion']}: {url}")
            run_path = f"{repo_path}/actions/runs/{run['id']}"
            artifacts = json.loads(api(f"{run_path}/artifacts?name={ARTIFACT}", authorization))["artifacts"]
            for artifact in artifacts:
                if artifact["name"] == ARTIFACT and not artifact["expired"]:
                    content = api(f"{repo_path}/actions/artifacts/{artifact['id']}/zip", authorization)
                    verify_artifact(content, sha)
                    print(f"Delivered {ARTIFACT}: {url}", flush=True)
                    return
            raise ValueError(f"Actions succeeded without a retained {ARTIFACT} artifact: {url}")
        time.sleep(3)
    raise TimeoutError(f"Timed out waiting for the local Actions run for {branch} at {sha}.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ref", default="HEAD", help="committed revision to seed (default HEAD)")
    parser.add_argument("--branch", help="destination branch (default source branch or local-ci)")
    parser.add_argument("--wait", action="store_true", help="wait for CI and verify the downloaded artifact")
    parser.add_argument("--timeout", type=int, default=1200, help="seconds to wait for CI (default 1200)")
    args = parser.parse_args()
    name = os.environ["LOCAL_CI_REPO"]
    if re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_-]*", name) is None:
        parser.error("LOCAL_CI_REPO must be a repository name containing letters, digits, dashes or underscores")
    if args.timeout <= 0:
        parser.error("--timeout must be positive")
    try:
        sha, branch = resolve_revision(args.ref, args.branch)
        if args.wait:
            git("cat-file", "-e", f"{sha}:{WORKFLOW}")
        password = Path("/credentials/password").read_text(encoding="ascii").strip()
        authorization = "Basic " + base64.b64encode(f"{OWNER}:{password}".encode("ascii")).decode("ascii")
        url = ensure_repository(name, branch, authorization)
        git("-c", "push.followTags=false", "-c", "push.gpgSign=false", "-c", "push.recurseSubmodules=no",
            "push", f"{SERVER}/{OWNER}/{name}.git", f"{sha}:refs/heads/{branch}", authorization=authorization)
        print(f"Seeded {branch} at {sha}. Actions: {url}/actions", flush=True)
        if args.wait:
            wait_for_delivery(name, sha, branch, authorization, args.timeout)
    except (ValueError, OSError, subprocess.CalledProcessError, zipfile.BadZipFile) as err:
        print(f"Local CI: {err}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
