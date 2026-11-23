"""Exercise Ctrl+C during a Git read in an isolated console or pseudo-terminal."""

import errno
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time

from session_console import check_session


def wait_for_git(ready):
    deadline = time.monotonic() + 5
    while time.monotonic() < deadline:
        try:
            pid = ready.read_text(encoding="ascii").strip()
        except FileNotFoundError:
            pid = ""
        if pid.isdecimal():
            return int(pid)
        time.sleep(0.02)
    raise AssertionError("Git subprocess did not become ready")


def run_posix(binary, env, ready):
    import pty
    import select

    pid, master = pty.fork()
    if pid == 0:
        os.chdir(ready.parent)
        os.execve(binary, [str(binary), "status"], env)
    reaped = False
    try:
        git_pid = wait_for_git(ready)
        os.write(master, b"\x03")
        output = bytearray()
        deadline = time.monotonic() + 3
        while time.monotonic() < deadline:
            readable, _, _ = select.select([master], [], [], 0.05)
            if readable:
                try:
                    output += os.read(master, 4096)
                except OSError as err:
                    if err.errno != errno.EIO:
                        raise
            done, status = os.waitpid(pid, os.WNOHANG)
            if done:
                reaped = True
                break
        if not reaped:
            raise AssertionError("git-view did not exit within three seconds of Ctrl+C")
        try:
            os.kill(git_pid, 0)
        except ProcessLookupError:
            return os.waitstatus_to_exitcode(status), bytes(output)
        raise AssertionError("Git subprocess survived Ctrl+C")
    finally:
        os.close(master)
        if not reaped:
            os.kill(pid, signal.SIGKILL)
            os.waitpid(pid, 0)


def run_windows(binary, env, ready):
    import ctypes
    from ctypes import wintypes

    kernel = ctypes.WinDLL("kernel32", use_last_error=True)
    kernel.OpenProcess.argtypes = [wintypes.DWORD, wintypes.BOOL, wintypes.DWORD]
    kernel.OpenProcess.restype = wintypes.HANDLE
    kernel.WaitForSingleObject.argtypes = [wintypes.HANDLE, wintypes.DWORD]
    kernel.WaitForSingleObject.restype = wintypes.DWORD
    kernel.CloseHandle.argtypes = [wintypes.HANDLE]
    startup = subprocess.STARTUPINFO()
    startup.dwFlags |= subprocess.STARTF_USESHOWWINDOW
    startup.wShowWindow = subprocess.SW_HIDE
    with subprocess.Popen(
        [binary, "status"],
        cwd=ready.parent,
        env=env,
        creationflags=subprocess.CREATE_NEW_CONSOLE,
        startupinfo=startup,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    ) as process:
        git = None
        try:
            git_pid = wait_for_git(ready)
            synchronize = 0x00100000
            git = kernel.OpenProcess(synchronize, False, git_pid)
            if not git:
                raise ctypes.WinError(ctypes.get_last_error())
            kernel.FreeConsole()
            if not kernel.AttachConsole(process.pid):
                raise ctypes.WinError(ctypes.get_last_error())
            try:
                if not kernel.SetConsoleCtrlHandler(None, True):
                    raise ctypes.WinError(ctypes.get_last_error())
                # Ctrl+C goes to every process in this new console only.
                if not kernel.GenerateConsoleCtrlEvent(0, 0):
                    raise ctypes.WinError(ctypes.get_last_error())
                stdout, stderr = process.communicate(timeout=3)
            finally:
                kernel.FreeConsole()
            if stdout:
                raise AssertionError(f"unexpected stdout after cancellation: {stdout!r}")
            if kernel.WaitForSingleObject(git, 0) != 0:
                raise AssertionError("Git subprocess survived Ctrl+C")
            return process.returncode, stderr
        finally:
            if process.poll() is None:
                process.kill()
            if git:
                kernel.CloseHandle(git)


def main():
    root = Path(__file__).resolve().parents[2]
    windows = os.name == "nt"
    suffix = ".exe" if windows else ""
    with tempfile.TemporaryDirectory(prefix="git-view-terminal-") as directory:
        work = Path(directory)
        binary = work / ("git-view" + suffix)
        git = work / ("git" + suffix)
        for output, package in (
            (binary, "./cmd/git-view"),
            (git, "./tests/acceptance/testdata/slowgit"),
        ):
            subprocess.run(["go", "build", "-o", output, package], cwd=root, check=True)
        check_session(binary, work)
        ready = work / "ready"
        env = dict(os.environ, PATH=str(work), GIT_VIEW_TEST_READY=str(ready))
        run = run_windows if windows else run_posix
        code, output = run(binary, env, ready)
        if code != 130 or b"context canceled" not in output or b"Repository:" in output:
            raise AssertionError(f"expected exit 130 without status output, got {code}: {output!r}")
    print("Ctrl+C returned exit 130 and stopped the immediate Git subprocess")


if __name__ == "__main__":
    main()
