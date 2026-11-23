package acceptance

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func TestLocalPipeline(t *testing.T) {
	if os.Getenv("GIT_VIEW_LOCAL_CI_TEST") != "1" {
		t.Skip("set GIT_VIEW_LOCAL_CI_TEST=1 to start an isolated Gitea pipeline")
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	project := "git-view-ci-test-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	compose := func(args ...string) (string, error) {
		cmd := exec.Command("docker", append([]string{"compose", "--project-name", project}, args...)...)
		cmd.Dir = "../.."
		cmd.Env = append(os.Environ(), "LOCAL_CI_PORT="+strconv.Itoa(port), "LOCAL_CI_SOURCE=.", "LOCAL_CI_REPO=git-view-cli")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("docker compose %v: %w\n%s", args, err, out)
		}
		return string(out), nil
	}
	var result string
	ref := os.Getenv("GIT_VIEW_LOCAL_CI_REF")
	if ref == "" {
		ref = "HEAD"
	}
	seed := []string{"run", "--build", "--rm", "bootstrap", "--ref", ref, "--branch", "acceptance", "--wait"}
	suite := godog.TestSuite{
		Name: "local-ci", Options: &godog.Options{
			Format: "pretty", Paths: []string{"../features/local_ci.feature"}, Strict: true, TestingT: t,
		},
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Step(`^the local Compose pipeline is configured$`, func() error {
				_, err := compose("config", "--quiet")
				if err == nil {
					t.Cleanup(func() {
						if t.Failed() {
							output, _ := compose("logs", "--no-color", "--tail", "20", "git", "provision", "runner")
							t.Log(output)
						}
						if _, err := compose("down", "--volumes"); err != nil {
							t.Error(err)
						}
					})
				}
				return err
			})
			sc.Step(`^I bootstrap a committed checkout and wait for its Actions run$`, func() error {
				var err error
				result, err = compose(seed...)
				return err
			})
			sc.Step(`^the pushed commit succeeds and delivers a candidate artifact$`, func() error {
				if !strings.Contains(result, "Delivered git-view-candidate") {
					return fmt.Errorf("expected a verified artifact on the server:\n%s", result)
				}
				return nil
			})
			sc.Step(`^repeating the bootstrap verifies the same run$`, func() error {
				again, err := compose(seed...)
				if err != nil {
					return err
				}
				_, delivered, _ := strings.Cut(result, "Delivered git-view-candidate: ")
				_, repeated, _ := strings.Cut(again, "Delivered git-view-candidate: ")
				deliveredURL, _, _ := strings.Cut(delivered, "\n")
				repeatedURL, _, _ := strings.Cut(repeated, "\n")
				if strings.TrimSpace(deliveredURL) != strings.TrimSpace(repeatedURL) {
					return fmt.Errorf("repeat bootstrap did not verify the same run:\n%s", again)
				}
				return nil
			})
		},
	}
	if suite.Run() != 0 {
		t.Fatal("local pipeline acceptance failed")
	}
}
