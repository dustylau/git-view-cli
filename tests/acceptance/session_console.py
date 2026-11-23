"""Exercise Interactive, Watch and foreground input using a native terminal."""

from contextlib import ExitStack
import ctypes
import json
import os
from pathlib import Path
import subprocess
import sys
import time


def exercise(send, expect, repo):
    expect('Branch: "main" (unborn)')
    send("files")
    expect("Empty directory.")
    send("watch --views changes")
    expect("No changes.")
    (repo / "space note").write_text("new", encoding="utf-8")
    expect('? ? "space note"')
    send("pause")
    expect("Watch paused.")
    send("resume")
    expect('? ? "space note"')
    send("back")
    expect('untracked file "space note"')
    send('edit -- "space note"')
    expect("Editor ready")
    send("foreground input")
    expect("Editor received foreground input")
    expect('untracked file "space note"')
    send('stage -- "space note"')
    expect("Proceed? [y/N]")
    send("y")
    expect('tracked file "space note"')
    send("quit")


def posix_session(binary, repo, env):
    import pty
    import select
    import termios

    master, slave = pty.openpty()
    try:
        before = termios.tcgetattr(slave)
        with subprocess.Popen([binary, "interactive"], cwd=repo, env=env,
                              stdin=slave, stdout=slave, stderr=slave) as child:
            pending = b""

            def expect(value):
                nonlocal pending
                needle = value.encode()
                deadline = time.monotonic() + 15
                while needle not in pending:
                    if time.monotonic() >= deadline or child.poll() is not None:
                        raise AssertionError(f"waiting for {value}: {pending!r}")
                    if select.select([master], [], [], 0.1)[0]:
                        pending += os.read(master, 8192)
                pending = pending.split(needle, 1)[1]

            try:
                exercise(lambda s: os.write(master, (s + "\n").encode()), expect, repo)
                assert child.wait(timeout=5) == 0
                assert termios.tcgetattr(slave) == before, "terminal attributes changed"
            finally:
                if child.poll() is None:
                    child.kill()
    finally:
        os.close(slave)
        os.close(master)


def windows_session(binary, repo, env):
    from ctypes import wintypes as wt
    import msvcrt

    kernel = ctypes.WinDLL("kernel32", use_last_error=True)

    class Coord(ctypes.Structure):
        _fields_ = [("x", wt.SHORT), ("y", wt.SHORT)]

    class BufferInfo(ctypes.Structure):
        _fields_ = [("size", Coord), ("cursor", Coord), ("attributes", wt.WORD),
                    ("window", wt.SMALL_RECT), ("maximum", Coord)]

    class Key(ctypes.Structure):
        _fields_ = [("down", wt.BOOL), ("repeat", wt.WORD), ("code", wt.WORD),
                    ("scan", wt.WORD), ("char", wt.WCHAR), ("control", wt.DWORD)]

    class Event(ctypes.Structure):
        _fields_ = [("type", wt.WORD), ("key", Key)]

    def check(ok):
        if not ok:
            raise ctypes.WinError(ctypes.get_last_error())

    with ExitStack() as stack:
        cin = os.open("CONIN$", os.O_RDWR)
        stack.callback(os.close, cin)
        cout = os.open("CONOUT$", os.O_RDWR)
        stack.callback(os.close, cout)
        ih = wt.HANDLE(msvcrt.get_osfhandle(cin))
        oh = wt.HANDLE(msvcrt.get_osfhandle(cout))
        mode = wt.DWORD()
        check(kernel.GetConsoleMode(ih, ctypes.byref(mode)))
        with subprocess.Popen([binary, "interactive"], cwd=repo, env=env,
                              stdin=cin, stdout=cout, stderr=cout) as child:
            offset = 0

            def send(value):
                events = []
                for c in value + "\r":
                    code = 13 if c == "\r" else 0
                    events.extend(Event(1, Key(down, 1, code, 0, c, 0)) for down in (1, 0))
                records = (Event * len(events))(*events)
                written = wt.DWORD()
                check(kernel.WriteConsoleInputW(ih, records, len(records), ctypes.byref(written)))

            def expect(value):
                nonlocal offset
                deadline = time.monotonic() + 15
                while True:
                    info = BufferInfo()
                    check(kernel.GetConsoleScreenBufferInfo(oh, ctypes.byref(info)))
                    count = info.size.x * (info.cursor.y + 1)
                    buf = ctypes.create_unicode_buffer(count)
                    n = wt.DWORD()
                    check(kernel.ReadConsoleOutputCharacterW(oh, buf, count, Coord(0, 0), ctypes.byref(n)))
                    screen = buf[:n.value]
                    found = screen.find(value, offset)
                    if found >= 0:
                        offset = found + len(value)
                        return
                    if child.poll() is not None or time.monotonic() >= deadline:
                        raise AssertionError(f"waiting for {value}: {screen[offset:]!r}")
                    time.sleep(0.05)

            try:
                exercise(send, expect, repo)
                assert child.wait(timeout=5) == 0
                after = wt.DWORD()
                check(kernel.GetConsoleMode(ih, ctypes.byref(after)))
                assert mode.value == after.value, "console input mode changed"
            finally:
                if child.poll() is None:
                    child.kill()


def check_session(binary, directory):
    repo = directory / "session-repo"
    repo.mkdir()
    env = dict(os.environ, GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_NOSYSTEM="1",
               GIT_VIEW_EDITOR=json.dumps([sys.executable, str(Path(__file__).resolve()), "--editor"]))
    subprocess.run(["git", "init", "-q", "-b", "main"], cwd=repo, env=env, check=True)
    if os.name == "nt":
        startup = subprocess.STARTUPINFO()
        startup.dwFlags |= subprocess.STARTF_USESHOWWINDOW
        startup.wShowWindow = subprocess.SW_HIDE
        result = subprocess.run([sys.executable, __file__, "--console", binary, repo], env=env,
                       creationflags=subprocess.CREATE_NEW_CONSOLE, startupinfo=startup,
                       stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=90)
        if result.returncode:
            raise AssertionError(result.stdout.decode(errors="replace"))
    else:
        posix_session(binary, repo, env)
    indexed = subprocess.check_output(["git", "ls-files", "-z"], cwd=repo, env=env)
    assert indexed == b"space note\0"
    print("Native session refreshed, returned to files, handed off input and confirmed staging")


if __name__ == "__main__":
    if sys.argv[1] == "--editor":
        print("Editor ready", flush=True)
        print("Editor received " + input(), flush=True)
    else:
        windows_session(Path(sys.argv[2]), Path(sys.argv[3]), os.environ.copy())
