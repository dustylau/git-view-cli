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
	python, prefix := hostPython()
	output := t.TempDir()
	var candidate string
	var native []byte
	var checksums []string
	suite := godog.TestSuite{
		Name: "release", Options: &godog.Options{
			Format: "pretty", Paths: []string{"../features/release.feature"}, Strict: true, TestingT: t,
		},
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Step(`^I package version "([^"]+)"$`, func(version string) error {
				candidate = filepath.Join(output, "git-view-"+version)
				args := append(slices.Clone(prefix), "../../tools/release/package_release.py",
					"--version", version, "--out", output)
				if out, err := exec.Command(python, args...).CombinedOutput(); err != nil {
					return fmt.Errorf("packaging failed: %w\n%s", err, out)
				}
				return nil
			})
			sc.Step(`^the candidate archives cover exactly:$`, func(table *godog.Table) error {
				entries, err := os.ReadDir(candidate)
				if err != nil {
					return err
				}
				wanted := []string{"SHA256SUMS"}
				checksums = nil
				for _, row := range table.Rows[1:] {
					osName, arch := row.Cells[0].Value, row.Cells[1].Value
					name := filepath.Base(candidate) + "-" + osName + "-" + arch + ".zip"
					wanted = append(wanted, name)
					archive, err := os.ReadFile(filepath.Join(candidate, name))
					if err != nil {
						return err
					}
					checksums = append(checksums, fmt.Sprintf("%x  %s", sha256.Sum256(archive), name))
					binary, err := inspectArchive(archive, osName, arch)
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
			sc.Step(`^the host executable reports version "([^"]+)"$`, func(version string) error {
				if len(native) == 0 {
					return fmt.Errorf("no candidate executable for host %s/%s", runtime.GOOS, runtime.GOARCH)
				}
				binary := filepath.Join(output, executableName(runtime.GOOS))
				if err := os.WriteFile(binary, native, 0o700); err != nil {
					return err
				}
				cmd := exec.Command(binary, "--version")
				cmd.Dir = output
				out, err := cmd.CombinedOutput()
				if err != nil || string(out) != "git-view "+version+"\n" {
					return fmt.Errorf("candidate version output: %q (%v)", out, err)
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

func inspectArchive(data []byte, osName, arch string) ([]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	if len(z.File) != 2 {
		return nil, fmt.Errorf("expected two archive entries, got %d", len(z.File))
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
