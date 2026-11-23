package acceptance

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"debug/elf"
	"debug/pe"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

func TestReleasePackaging(t *testing.T) {
	if os.Getenv("GIT_VIEW_RELEASE_TEST") != "1" {
		t.Skip("set GIT_VIEW_RELEASE_TEST=1 to build and inspect the candidate archives")
	}
	output := t.TempDir()
	version := os.Getenv("GIT_VIEW_RELEASE_VERSION")
	if version == "" {
		version = "0.1.0-rc.1"
	}
	var candidate string
	var native []byte
	var checksums []string
	installed := &fixture{binary: filepath.Join(output, executableName(runtime.GOOS)), cwd: output, env: cleanEnv()}
	suite := godog.TestSuite{
		Name: "release", Options: &godog.Options{
			Format: "pretty", Paths: []string{"../features/release.feature"}, Strict: true, TestingT: t,
		},
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Step(`^I select the release candidate$`, func() error {
				if supplied := os.Getenv("GIT_VIEW_RELEASE_DIR"); supplied != "" {
					candidate = supplied
					return nil
				}
				candidate = filepath.Join(output, "git-view-"+version)
				python, prefix := hostPython()
				args := append(slices.Clone(prefix), "../../tools/release/package_release.py",
					"--version", version, "--out", output)
				if out, err := exec.Command(python, args...).CombinedOutput(); err != nil {
					return fmt.Errorf("packaging failed: %w\n%s", err, out)
				}
				return nil
			})
			sc.Step(`^the candidate archives include the Go notices and cover exactly:$`, func(table *godog.Table) error {
				notices, err := installedGoNotices()
				if err != nil {
					return err
				}
				entries, err := os.ReadDir(candidate)
				if err != nil {
					return err
				}
				wanted := []string{"SHA256SUMS"}
				checksums = nil
				for _, row := range table.Rows[1:] {
					osName, arch := row.Cells[0].Value, row.Cells[1].Value
					name := "git-view-" + version + "-" + osName + "-" + arch + ".zip"
					wanted = append(wanted, name)
					archive, err := os.ReadFile(filepath.Join(candidate, name))
					if err != nil {
						return err
					}
					checksums = append(checksums, fmt.Sprintf("%x  %s", sha256.Sum256(archive), name))
					binary, err := inspectArchive(archive, osName, arch, notices)
					if err != nil {
						return fmt.Errorf("%s: %w", name, err)
					}
					if osName == runtime.GOOS && arch == runtime.GOARCH {
						native = binary
					}
				}
				slices.Sort(wanted)
				var actual []string
				for _, entry := range entries {
					actual = append(actual, entry.Name())
				}
				if !reflect.DeepEqual(actual, wanted) {
					return fmt.Errorf("candidate files %v, want %v", actual, wanted)
				}
				return nil
			})
			sc.Step(`^the checksum manifest verifies every archive$`, func() error {
				data, err := os.ReadFile(filepath.Join(candidate, "SHA256SUMS"))
				if err != nil {
					return err
				}
				if string(data) != strings.Join(checksums, "\n")+"\n" {
					return fmt.Errorf("checksum manifest differs from the four calculated hashes")
				}
				return nil
			})
			sc.Step(`^the host executable reports the requested version$`, func() error {
				if len(native) == 0 {
					return fmt.Errorf("no candidate executable for host %s/%s", runtime.GOOS, runtime.GOARCH)
				}
				if err := os.WriteFile(installed.binary, native, 0o700); err != nil {
					return err
				}
				if err := installed.run("version"); err != nil {
					return err
				}
				if installed.code != 0 || installed.err.Len() != 0 || installed.out.String() != "git-view "+version+"\n" {
					return fmt.Errorf("candidate version: exit %d, stdout %q, stderr %q", installed.code, &installed.out, &installed.err)
				}
				return nil
			})
			sc.Step(`^the host executable inspects a mixed-change worktree without changing it$`, func() error {
				if err := installed.create(t, "mixed"); err != nil {
					return err
				}
				for _, check := range []struct {
					command string
					lines   []string
				}{
					{"status", []string{`Branch: "main"`, "State: dirty", "Changes: staged 1, unstaged 2, untracked 1, conflicts 0"}},
					{"changes", []string{`M M "tracked.txt"`, `. D "deleted.txt"`, `? ? "-new Ω file.txt"`}},
				} {
					if err := installed.run(check.command); err != nil {
						return err
					}
					if installed.code != 0 || installed.err.Len() != 0 {
						return fmt.Errorf("candidate %s: exit %d, stderr %q", check.command, installed.code, &installed.err)
					}
					for _, line := range check.lines {
						if !strings.Contains(installed.out.String(), line+"\n") {
							return fmt.Errorf("candidate %s: expected %q in %q", check.command, line, &installed.out)
						}
					}
					after, err := snapshot(installed.root)
					if err != nil {
						return err
					}
					if !reflect.DeepEqual(after, installed.before) {
						return fmt.Errorf("candidate %s changed the Git fixture", check.command)
					}
				}
				return nil
			})
		},
	}
	if suite.Run() != 0 {
		t.Fatal("release packaging acceptance failed")
	}
}

func executableName(osName string) string {
	if osName == "windows" {
		return "git-view.exe"
	}
	return "git-view"
}

func installedGoNotices() (map[string][]byte, error) {
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return nil, fmt.Errorf("resolve Go toolchain: %w", err)
	}
	root := strings.TrimSpace(string(out))
	notices := make(map[string][]byte)
	for _, name := range []string{"LICENSE", "PATENTS"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		notices["licenses/go/"+name] = data
	}
	return notices, nil
}

func inspectArchive(data []byte, osName, arch string, notices map[string][]byte) ([]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for name, want := range notices {
		got, err := fs.ReadFile(z, name)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(got, want) {
			return nil, fmt.Errorf("%s differs from the Go toolchain notice", name)
		}
	}
	if count := len(notices) + 2; len(z.File) != count {
		return nil, fmt.Errorf("expected %d archive entries, got %d", count, len(z.File))
	}
	if _, err := fs.Stat(z, "INSTALL.md"); err != nil {
		return nil, err
	}
	name := executableName(osName)
	file, err := z.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if osName == "linux" {
		info, err := file.Stat()
		if err != nil {
			return nil, err
		}
		if info.Mode().Perm() != 0o755 {
			return nil, fmt.Errorf("Linux executable mode is %o", info.Mode().Perm())
		}
	}
	binary, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if osName == "windows" {
		f, err := pe.NewFile(bytes.NewReader(binary))
		if err != nil {
			return nil, err
		}
		want := map[string]uint16{"amd64": pe.IMAGE_FILE_MACHINE_AMD64, "arm64": pe.IMAGE_FILE_MACHINE_ARM64}[arch]
		if f.Machine != want {
			return nil, fmt.Errorf("PE architecture %#x, want %#x", f.Machine, want)
		}
	} else {
		f, err := elf.NewFile(bytes.NewReader(binary))
		if err != nil {
			return nil, err
		}
		want := map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}[arch]
		if f.Machine != want {
			return nil, fmt.Errorf("ELF architecture %v, want %v", f.Machine, want)
		}
	}
	return binary, nil
}
