# ADR 0006: Line-oriented sessions and Git actions

Status: Accepted for 1.0.0

Use normal terminal line input and a single background Watch refresh at a time. A two-second reconciliation reads Git and the worktree on every platform, including linked worktrees. This provides a bounded refresh cadence on filesystems with inconsistent event delivery. Consider native event hints if measurements justify the additional dependency.

Use Go's terminal package for descriptor checks. Leave cooked input, scrolling and resize to the terminal. Interactive accepts the public command grammar and retains the active view when users return from a tool or Watch. A configured foreground child inherits the terminal while the renderer is idle.

Keep Git responsible for patches, revision lookup, mergetool and index transactions. Validate literal root-relative file targets before handing off. Editor configuration uses an argument array to support wait flags and paths containing spaces. The MIT project license accompanies the Go distribution notices.
