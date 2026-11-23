package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gitview "git-view-cli/internal/git"
)

func renderView(ctx context.Context, w gitview.Workspace, s gitview.Snapshot, r request) (string, error) {
	file := ""
	if len(r.paths) > 0 {
		file = r.paths[0]
	}
	var out strings.Builder
	switch r.name {
	case "status":
		return renderStatus(s), nil
	case "changes":
		return renderChanges(s), nil
	case "files":
		if file == "" {
			file = "."
		}
		entries, err := w.Files(ctx, file)
		if err != nil {
			return "", err
		}
		for _, e := range entries {
			fmt.Fprintf(&out, "%s %s %s\n", e.State, e.Kind, strconv.QuoteToGraphic(e.Path))
		}
		if len(entries) == 0 {
			return "Empty directory.\n", nil
		}
	case "log", "history":
		commits, err := w.Log(ctx, s.Branch, file, r.limit, r.skip)
		if err != nil {
			return "", err
		}
		for _, c := range commits {
			fmt.Fprintf(&out, "%s  %s  %s  %s\n", strconv.QuoteToGraphic(c.OID), strconv.QuoteToGraphic(c.Time), strconv.QuoteToGraphic(c.Author), strconv.QuoteToGraphic(c.Subject))
		}
		if len(commits) == 0 {
			return "No commits.\n", nil
		}
	case "diff", "show":
		if !r.staged && file != "" {
			for _, c := range s.Changes {
				if c.Path == file && c.Untracked {
					return "Untracked file " + strconv.QuoteToGraphic(file) + "\n", nil
				}
			}
		}
		data, large, err := w.Patch(ctx, r.revision, file, r.staged)
		if err != nil {
			return "", err
		}
		if large {
			return "Patch exceeds the display limit (256 KiB). Narrow the diff to a file.\n", nil
		}
		if len(data) == 0 {
			return "No differences.\n", nil
		}
		return renderPatch(data), nil
	case "conflicts":
		entries, err := w.Conflicts(ctx)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			fmt.Fprint(&out, strconv.QuoteToGraphic(entry.Path))
			for i, label := range []string{"", "base", "ours", "theirs"} {
				if oid := entry.Stages[i]; oid != "" {
					fmt.Fprintf(&out, "  %s %s", label, strconv.QuoteToGraphic(oid))
				}
			}
			out.WriteByte('\n')
		}
		if len(entries) == 0 {
			return "No conflicts.\n", nil
		}
	default:
		return "", fmt.Errorf("unknown view %q", r.name)
	}
	return out.String(), nil
}

func renderPatch(data []byte) string {
	var out strings.Builder
	for line := range strings.SplitSeq(strings.TrimSuffix(string(data), "\n"), "\n") {
		escaped := strconv.QuoteToGraphic(line)
		out.WriteString(escaped[1 : len(escaped)-1])
		out.WriteByte('\n')
	}
	return out.String()
}
