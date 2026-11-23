"""Build and test git-view in Docker using a seeded volume or a host checkout."""

import argparse
import csv
import io
from pathlib import Path
import subprocess
import sys
import tempfile
import uuid


ROOT = Path(__file__).resolve().parents[2]
IMAGE_TAG = "git-view:test"
SCRIPT_DIR = "/src/tests/container"


def docker(*args, check=True):
    return subprocess.run(["docker", *args], cwd=ROOT, check=check)


def mount(*fields):
    # Docker parses --mount as CSV. Quote fields such as paths containing commas.
    buffer = io.StringIO()
    csv.writer(buffer, lineterminator="\r\n").writerow(fields)
    return buffer.getvalue()[:-2]


def build(invoke):
    with tempfile.TemporaryDirectory(prefix="git-view-build-") as directory:
        image_file = Path(directory) / "image-id"
        invoke("build", "--load", "--tag", IMAGE_TAG, "--iidfile", str(image_file), ".")
        return image_file.read_text(encoding="ascii").strip()


def run_container(invoke, image, repository_mount, script, *args):
    invoke(
        "run", "--rm", "--init", "--network=none", "--mount", repository_mount,
        image, "sh", f"{SCRIPT_DIR}/{script}", *args,
    )


def run_workflow(repo=None, invoke=docker):
    if repo is not None:
        repo = repo.resolve()
        if not repo.is_dir():
            raise ValueError(f"repository directory does not exist: {repo}")

    image = build(invoke)
    if repo is not None:
        repository_mount = mount(
            "type=bind", f"source={repo}", "target=/repo", "readonly",
            "bind-recursive=disabled",
        )
        run_container(invoke, image, repository_mount, "check.sh", "host")
        return

    volume = f"git-view-fixture-{uuid.uuid4().hex}"
    invoke("volume", "create", "--label", "git-view.purpose=container-test", volume)
    print(f"Fixture volume: {volume}", flush=True)
    fields = ("type=volume", f"source={volume}", "target=/repo", "volume-nocopy")
    completed = False
    try:
        run_container(invoke, image, mount(*fields), "seed.sh")
        run_container(invoke, image, mount(*fields, "readonly"), "check.sh", "seeded")
        completed = True
    finally:
        cleanup = invoke("volume", "rm", volume, check=False)
        if cleanup.returncode:
            print(f"Could not remove fixture volume {volume}.", file=sys.stderr)
            if completed:
                cleanup.check_returncode()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", type=Path, help="mount this host checkout read-only")
    args = parser.parse_args()
    try:
        run_workflow(repo=args.repo)
    except subprocess.CalledProcessError as err:
        print(f"Docker command failed with exit code {err.returncode}.", file=sys.stderr)
        return err.returncode
    except OSError as err:
        print(f"Could not run the container workflow: {err}", file=sys.stderr)
        return 1
    except ValueError as err:
        parser.error(str(err))
    except KeyboardInterrupt:
        return 130
    return 0


if __name__ == "__main__":
    sys.exit(main())
