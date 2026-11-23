# Local CI and delivery

Start Docker with Linux containers and Docker Compose v2 or later. From a normal checkout of this repository, run the bootstrapper for your shell.

```powershell
.\tools\ci\bootstrap.ps1 --wait
```

```sh
sh tools/ci/bootstrap.sh --wait
```

The bootstrapper builds the tools image, starts Gitea and its runner, creates a private `local-ci/git-view-cli` repository, and pushes your committed HEAD. It uses the current branch name, or `local-ci` for a detached HEAD. It carries only the selected commit and its history. Local remotes and uncommitted work stay in the checkout. Use `--ref` to choose a committed revision or `--branch` to choose its destination. An existing branch must advance by fast-forward.

After a successful push, `--wait` waits for its Actions run, downloads its artifact for checksum verification, and prints the run page URL. Omit `--wait` to return after the push. Pushing an unchanged commit reuses the existing run. Use the Actions UI to rerun a failed job or refresh expired artifacts.

## Run history and downloads

Open the URL printed by the bootstrapper, or visit `http://127.0.0.1:3080` and select the repository's **Actions** tab. The local username is `local-ci`. Read its generated password from the Compose volume in your terminal:

```sh
docker compose run --rm --no-deps --entrypoint cat bootstrap /credentials/password
```

Each successful run offers a `git-view-candidate` artifact containing the four MVP ZIPs and `SHA256SUMS`. The filenames and executable version identify the triggering commit. See the [installation guide](INSTALL.md) for extracting and running a ZIP. For subsequent commits, rerun the bootstrapper or push to the HTTP clone URL shown in Gitea.

The infrastructure, seeding, compilation, job execution, tests, and packaging run in containers. The host needs Docker Compose and a shell. First use downloads images, public GitHub Actions, and build dependencies. GitHub authentication is unnecessary. The runner uses the Docker daemon and is intended for trusted local workflows. The job containers have their own checked-out workspace.

## Defaults and overrides

| Setting                | Default        | When to change it                                                     |
| ---------------------- | -------------- | --------------------------------------------------------------------- |
| `LOCAL_CI_PORT`        | `3080`         | The loopback port is in use.                                          |
| `LOCAL_CI_REPO`        | `git-view-cli` | You want a separate private repository for another checkout.          |
| `LOCAL_CI_SOURCE`      | `.`            | Seed another ordinary clone by supplying its absolute directory path. |
| `COMPOSE_PROJECT_NAME` | `git-view-ci`  | Run a separate stack with its own port and volumes.                   |

Set overrides in your shell or Compose `.env`. The source path must be mountable by Docker and contain the Git metadata. Use an ordinary clone for repositories whose linked worktree Git directory sits outside the checkout. The committed revision needs the local workflow for `--wait` to work.

Use one active CI stack per Docker daemon. Switch between checkouts in the same stack with `LOCAL_CI_SOURCE` and `LOCAL_CI_REPO`.

Use `docker compose ps -a` for service state and `docker compose logs git provision runner` for startup failures. `docker compose stop` stops the stack. Run the bootstrapper again to restart it with the same Git history, credentials, artifacts, and runner registration.

Local CI executes the Linux Docker architecture and cross-builds the other artifact targets. The [hosted workflows](../.github/workflows/go.yml) cover the separate native validation plan.
