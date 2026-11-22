# Security

## Trust boundaries

This planned application runs locally with the user's filesystem permissions. Repository paths, filenames, commit messages, branch names, file contents, Git configuration, and external tools can be untrusted inputs. No telemetry or remote service is planned.

## Required safeguards

- Launch Git with argument arrays, validated revisions, and literal filename/pathspec handling. Do not interpolate repository data into shell commands.
- Keep automatic reading separate from explicit external tools. Disable external diff/textconv execution for automatic views; configured editors, pagers, difftools, and mergetools can execute arbitrary user-configured programs.
- Escape terminal control characters in displayed metadata and content; prevent file contents from controlling the user's terminal. Bound content, subprocess output, and memory usage.
- Restrict navigation/tool paths to the chosen worktree after resolving symlinks and handle path replacement races. Do not follow links outside the root automatically.
- Do not change ownership/safe-directory settings, Git hooks, global config, credentials, or signing behavior. Report Git ownership rejection with context.
- Do not implicitly fetch, stage, commit, reset, discard, clean, or continue/abort a Git operation. Show explicit mutation targets and re-query state after authorized changes.
- Recognize that Git mergetool may stage changes or create backups according to Git configuration. Explain and report those effects.
- Keep diagnostics opt-in and avoid logging file contents, credentials, environment secrets, or complete command environments. Store temporary files securely and clean up children/resources.

## Reporting

No public security contact or reporting channel has been established. Define one before public release. Do not include credentials or private repository content in public reports. There is no release or supported-version security policy yet.
