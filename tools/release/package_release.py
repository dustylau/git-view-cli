"""Build and checksum local git-view candidate archives."""

import argparse
import hashlib
import os
from pathlib import Path
import re
import stat
import subprocess
import sys
import tempfile
import zipfile


ROOT = Path(__file__).resolve().parents[2]
TARGETS = (
    ("linux", "amd64"),
    ("linux", "arm64"),
    ("windows", "amd64"),
    ("windows", "arm64"),
)
VERSION = re.compile(
    r"(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)"
    r"(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?"
    r"(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?"
)


def go_environment():
    return dict(os.environ, CGO_ENABLED="0", GOFLAGS="", GOEXPERIMENT="", GOWORK="off")


def build_target(os_name, arch, version, binary):
    env = dict(go_environment(), GOOS=os_name, GOARCH=arch, GOAMD64="v1", GOARM64="v8.0")
    print(f"Building {os_name}/{arch}", flush=True)
    subprocess.run(
        [
            "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false",
            "-ldflags", f"-X git-view-cli/internal/cli.Version={version}",
            "-o", str(binary), "./cmd/git-view",
        ],
        cwd=ROOT, env=env, check=True,
    )


def distribution_documents():
    go_root = Path(subprocess.check_output(
        ["go", "env", "GOROOT"], cwd=ROOT, env=go_environment(), text=True, encoding="utf-8",
    ).strip())
    instructions = (ROOT / "docs/INSTALL.md").read_text(encoding="utf-8").encode("utf-8")
    documents = {"INSTALL.md": instructions}
    for name in ("LICENSE", "PATENTS"):
        notice = go_root / name
        contents = notice.read_bytes()
        if not contents.strip():
            raise OSError(f"Go toolchain notice is empty: {notice}")
        documents[f"licenses/go/{name}"] = contents
    return documents


def write_zip(archive, binary, documents):
    files = [(name, contents, 0o644) for name, contents in documents.items()]
    files.append((binary.name, binary.read_bytes(), 0o755))
    with zipfile.ZipFile(archive, "w") as target:
        for name, contents, mode in files:
            info = zipfile.ZipInfo(name)
            info.create_system = 3
            info.external_attr = (stat.S_IFREG | mode) << 16
            target.writestr(info, contents, compress_type=zipfile.ZIP_DEFLATED, compresslevel=9)


def package_release(version, output, build=build_target):
    if VERSION.fullmatch(version) is None:
        raise ValueError("use a version like 0.1.0-rc.1")
    stem = f"git-view-{version}"
    candidate = output.resolve() / stem
    if candidate.exists() or candidate.is_symlink():
        raise FileExistsError(f"candidate already exists: {candidate}")
    documents = distribution_documents()
    candidate.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=".git-view-build-", dir=candidate.parent) as directory:
        scratch = Path(directory)
        staged = scratch / stem
        staged.mkdir()
        checksums = []
        for os_name, arch in TARGETS:
            binary = scratch / ("git-view.exe" if os_name == "windows" else "git-view")
            build(os_name, arch, version, binary)
            filename = f"{stem}-{os_name}-{arch}.zip"
            archive = staged / filename
            write_zip(archive, binary, documents)
            checksums.append(f"{hashlib.sha256(archive.read_bytes()).hexdigest()}  {filename}\n")
        (staged / "SHA256SUMS").write_bytes("".join(checksums).encode("ascii"))
        if candidate.exists() or candidate.is_symlink():
            raise FileExistsError(f"candidate already exists: {candidate}")
        staged.rename(candidate)
    return candidate


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True, help="candidate version, for example 0.1.0-rc.1")
    parser.add_argument("--out", type=Path, default=ROOT / "dist", help="parent output directory")
    args = parser.parse_args()
    try:
        destination = package_release(args.version, args.out)
    except ValueError as err:
        parser.error(str(err))
    except subprocess.CalledProcessError as err:
        print(f"Go command failed with exit code {err.returncode}.", file=sys.stderr)
        return err.returncode
    except OSError as err:
        print(f"Could not package the candidate: {err}", file=sys.stderr)
        return 1
    print(f"Candidate saved to {destination}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
