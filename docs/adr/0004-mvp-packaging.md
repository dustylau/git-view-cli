# ADR 0004: Package MVP candidates as ZIP archives

Status: Accepted for local packaging

Use one ZIP per supported target plus a SHA-256 manifest. Both operating systems can extract ZIP archives. Keep the Linux executable mode in the archive. Include the installation instructions alongside each binary so the archive can be used outside the checkout.

Set the candidate version with the [Go linker's `-X` option](https://pkg.go.dev/cmd/link). Build with fixed architecture baselines and CGO disabled. Strip checkout paths and VCS stamping from the build. The package tool uses Python's standard library and gives each ZIP entry fixed metadata. The architecture matrix follows the MVP specification.

The package command stages all files before placing them in the destination. Native validation, license selection, and publication approval follow the release policy.

Build the candidate once for a native verification run. Pass it between jobs by its [GitHub Actions workflow artifact](https://docs.github.com/en/actions/tutorials/store-and-share-data) ID. Checking the same archive bytes on each target ties results to the candidate that was actually tested. Dispatch is explicit and the workflow's credentials have read-only repository access.
