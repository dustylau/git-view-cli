package acceptance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func hostPython() (string, []string) {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("py"); err == nil {
			return "py", []string{"-3"}
		}
		return "python", nil
	}
	return "python3", nil
}

func TestContainerWorkflow(t *testing.T) {
	if os.Getenv("GIT_VIEW_CONTAINER_TEST") != "1" {
		t.Skip("set GIT_VIEW_CONTAINER_TEST=1 to run the Docker integration scenarios")
	}
	python, prefix := hostPython()
	test := godog.TestSuite{
		Name: "container", Options: &godog.Options{
			Format: "pretty", Paths: []string{"../features/container.feature"}, Strict: true, TestingT: t,
		},
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			var host *fixture
			var output []byte
			var runErr error
			sc.Step(`^the container workflow uses a (seeded|host) fixture$`, func(mode string) error {
				host = nil
				if mode == "host" {
					host = &fixture{}
					return host.create(t, "mixed")
				}
				return nil
			})
			sc.Step(`^I run the container build and test workflow$`, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				defer cancel()
				args := append(append([]string{}, prefix...), "../container/container_workflow.py")
				if host != nil {
					args = append(args, "--repo", host.selected)
				}
				output, runErr = exec.CommandContext(ctx, python, args...).CombinedOutput()
			})
			sc.Step(`^the container workflow succeeds$`, func() error {
				if runErr != nil {
					return fmt.Errorf("workflow failed: %w\n%s", runErr, output)
				}
				return nil
			})
			sc.Step(`^the container output contains (".*")$`, func(quoted string) error {
				want, err := strconv.Unquote(quoted)
				if err != nil {
					return err
				}
				if !strings.Contains(string(output), want) {
					return fmt.Errorf("workflow output missing %q\n%s", want, output)
				}
				return nil
			})
			sc.Step(`^its seeded volume has been removed$`, func() error {
				volume := regexp.MustCompile(`Fixture volume: (git-view-fixture-[a-f0-9]+)`).FindSubmatch(output)
				if volume == nil {
					return fmt.Errorf("workflow did not identify its fixture volume")
				}
				out, err := exec.Command("docker", "volume", "ls", "-q", "--filter", "name="+string(volume[1])).CombinedOutput()
				if err != nil || strings.TrimSpace(string(out)) != "" {
					return fmt.Errorf("volume cleanup check: %v %s", err, out)
				}
				return nil
			})
			sc.Step(`^the host fixture is unchanged$`, func() error {
				after, err := snapshot(host.root)
				if err != nil {
					return err
				}
				if !reflect.DeepEqual(host.before, after) {
					return fmt.Errorf("container workflow changed the host checkout")
				}
				return nil
			})
		},
	}
	if test.Run() != 0 {
		t.Fatal("container workflow acceptance failed")
	}
}
