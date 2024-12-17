package browser

import (
	"context"
	"os/exec"
	"runtime"
)

func Open(url string) error {
	return OpenContext(context.Background(), url)
}

func OpenContext(ctx context.Context, url string) error {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "windows":
		cmd, args = "cmd", []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		// "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return err
	}
	args = append(args, url)
	return exec.CommandContext(ctx, cmd, args...).Start()
}
