# ADR 0005: Local CI and delivery with Gitea

Status: Accepted for local development

Use [Gitea Actions](https://docs.gitea.com/1.27/usage/actions/comparison/) to run Git-triggered CI in Docker. Gitea runs on loopback with SQLite and persistent Compose storage. A private repository starts from a committed revision of the developer's checkout. The bootstrap uses Git and the Gitea API inside a container without modifying the checkout.

Use `.gitea/workflows/` for the local Linux job and keep the hosted native matrix in `.github/workflows/`. The Gitea runner uses the local Docker daemon to start disposable job containers. Leave the socket and checkout mount out of job containers. The local instance is for trusted repositories and workflows.

Deliver versioned, verified ZIPs as an artifact attached to the successful local run. Use the pinned [Gitea-compatible v4 upload action](https://github.com/ChristopherHX/gitea-upload-artifact) so the artifact is accessible through both the run page and the REST API. Public releases follow the existing approval and native-validation gates.
