package util

import (
    "os/exec"
    "fmt"
)

func pkgManager(action, software string) error {

	var cmd *exec.Cmd

	switch action {
	case "install":
		// Command to install the software
		cmd = exec.Command("apt", "install", "-y", software)
	case "remove":
		// Command to uninstall the software
		cmd = exec.Command("apt", "uninstall", "-y", software)
	case "check":
		queryCmd := fmt.Sprintf("dpkg -l | grep ^ii | awk '{print $2}' | grep ^%s$", software)
		cmd = exec.Command("sh", "-c", queryCmd)
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing command: %v\nOutput: %s", err, output)
	}

	// If the action is "check" and there's no output, it means the software is not installed
	if action == "check" && len(output) == 0 {
		fmt.Printf("%s is not installed.\n", software)
	} else {
		fmt.Printf("%s\n", output)
	}

	return nil
}