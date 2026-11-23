# Install git-view from an archive

Choose the ZIP for your operating system and architecture. Windows and Linux each have an amd64 archive for x86-64 systems and an arm64 archive for ARM systems. The CLI uses Git 2.43.0 or newer for repository inspection. Running a packaged binary requires Git on PATH.

The `licenses/go/` directory contains the Go toolchain's license and patent grant.

## Windows

Run these commands from the folder containing the archive and `SHA256SUMS`. Replace the example archive name with the one you received.

```powershell
$archive = 'git-view-0.1.0-rc.1-windows-amd64.zip'
Get-FileHash -Algorithm SHA256 -LiteralPath $archive
Get-Content SHA256SUMS | Select-String -SimpleMatch $archive
```

Compare the two hexadecimal hashes. When they match, extract into a new directory and run the executable.

```powershell
Expand-Archive -LiteralPath $archive -DestinationPath .\git-view
.\git-view\git-view.exe --version
.\git-view\git-view.exe --repo 'D:\path\to\checkout' status
```

## Linux

Run these commands from the folder containing the archive and `SHA256SUMS`. Replace the example archive name with the one you received. Proceed if the checksum command reports that archive as `OK`.

```sh
archive=git-view-0.1.0-rc.1-linux-amd64.zip
sha256sum --ignore-missing --check SHA256SUMS
unzip "$archive" -d ./git-view
./git-view/git-view --version
./git-view/git-view --repo /path/to/checkout status
```

## Upgrade or roll back

Verify and extract each new version into a separate directory. Check `--version` and `status` before switching your PATH entry or launcher to the new directory. Retain the previous directory so you can switch back if needed.
