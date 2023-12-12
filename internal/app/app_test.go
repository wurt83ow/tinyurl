package app_test

import (
	"context"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestRunServerAndShutdown(t *testing.T) {
	// Set up a context with a timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a command to run your server
	cmd := exec.CommandContext(ctx, "../../cmd/shortener/shortener")

	// Create a wait group to wait for server shutdown
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the server in a separate goroutine
	go func() {
		defer wg.Done()
		if err := cmd.Run(); err != nil {
			t.Errorf("Error running server: %v", err)
		}
	}()

	// Allow some time for the server to start
	time.Sleep(1 * time.Second)

	// Send a SIGTERM signal to the server
	cmd.Process.Signal(syscall.SIGTERM)

	// Wait for the server to exit or the timeout to occur
	select {
	case <-time.After(30 * time.Second): // Use the same timeout as the test
		t.Error("Timeout waiting for server to exit")
	case <-ctx.Done():
	}

	// Wait for the server to exit
	wg.Wait()

	// Check if the server process exited successfully
	if err := cmd.Wait(); err != nil && !isChildProcessExitedSuccessfully(err) {
		t.Errorf("Error waiting for server to exit: %v", err)
	}
}

// Check if the error indicates that the child process exited successfully
func isChildProcessExitedSuccessfully(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		status, ok := exitErr.Sys().(syscall.WaitStatus)
		return ok && status.Exited() && status.ExitStatus() == 0
	}
	return false
}
