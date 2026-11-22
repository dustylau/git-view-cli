"""Measure warm git-view invocations and the equivalent Git status query."""

import math
import os
from pathlib import Path
import platform
import statistics
import subprocess
import tempfile
import time


WARMUPS = 3
SAMPLES = 20
FIXTURES = [("small", 100, 10, 400), ("medium", 5000, 1000, 1000)]


def run(command, cwd, env, timeout=30):
    result = subprocess.run(command, cwd=cwd, env=env, capture_output=True, timeout=timeout)
    if result.returncode:
        message = result.stderr.decode(errors="replace").strip()
        raise RuntimeError(f"{command[0]} exited {result.returncode}: {message}")
    return result.stdout


def create_fixture(repo, env, tracked, untracked):
    repo.mkdir()
    run(["git", "init", "--template=", "--initial-branch=main"], repo, env)
    (repo / "tracked").mkdir()
    (repo / "new").mkdir()
    for i in range(tracked):
        # Reuse a small set of blobs to keep fixture setup cheap on Windows.
        (repo / "tracked" / f"{i:05}.txt").write_bytes(f"original {i % 20}\n".encode())
    run(["git", "add", "--", "tracked"], repo, env, timeout=180)
    run(
        ["git", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
         "commit", "-m", "base"],
        repo, env, timeout=180,
    )
    for i in range(tracked // 10):
        (repo / "tracked" / f"{i:05}.txt").write_bytes(f"changed content {i}\n".encode())
    for i in range(untracked):
        (repo / "new" / f"{i:05}.txt").write_bytes(b"untracked\n")


def measure(commands, repo, env):
    samples = {name: [] for name in commands}
    # Interleave commands to keep the conditions comparable.
    for trial in range(WARMUPS + SAMPLES):
        for name, command in commands.items():
            start = time.perf_counter()
            run(command, repo, env)
            if trial >= WARMUPS:
                samples[name].append((time.perf_counter() - start) * 1000)
    return samples


def main():
    root = Path(__file__).resolve().parents[2]
    env = {key: value for key, value in os.environ.items() if not key.upper().startswith("GIT_")}
    env.update(
        GIT_CONFIG_NOSYSTEM="1",
        GIT_CONFIG_GLOBAL=os.devnull,
        GIT_NO_LAZY_FETCH="1",
        GIT_ALLOW_PROTOCOL="",
        GIT_TERMINAL_PROMPT="0",
    )
    print(platform.platform(), flush=True)
    for command in (["go", "version"], ["git", "--version"]):
        print(run(command, root, env).decode().strip(), flush=True)
    print(f"{WARMUPS} warmups and {SAMPLES} timed samples per command", flush=True)
    failures = []
    with tempfile.TemporaryDirectory(prefix="git-view-perf-") as directory:
        work = Path(directory)
        binary = work / ("git-view.exe" if os.name == "nt" else "git-view")
        run(["go", "build", "-buildvcs=false", "-o", binary, "./cmd/git-view"], root, env)
        commands = {
            "git": [
                "git", "--no-optional-locks", "-c", "core.fsmonitor=false",
                "-c", "core.untrackedCache=false", "status", "--porcelain=v2",
                "--branch", "--ahead-behind", "-z", "--untracked-files=all",
                "--renames", "--ignore-submodules=none", "--",
            ],
            "status": [binary, "status"],
            "changes": [binary, "changes"],
        }
        for label, tracked, untracked, budget in FIXTURES:
            repo = work / label
            create_fixture(repo, env, tracked, untracked)
            print(f"{label}: {tracked} tracked, {untracked} untracked", flush=True)
            for name, samples in measure(commands, repo, env).items():
                p95 = sorted(samples)[math.ceil(0.95 * len(samples)) - 1]
                print(f"  {name:7} median {statistics.median(samples):6.1f} ms, "
                      f"p95 {p95:6.1f} ms", flush=True)
                if name != "git" and p95 > budget:
                    failures.append(f"{label} {name} p95 {p95:.1f} ms exceeded {budget} ms")
    if failures:
        raise SystemExit("\n".join(failures))


if __name__ == "__main__":
    main()
