package util

import (
    "os/exec"
    "fmt"
	"log"
)

func detectPackageManager() (string, error) {

	log.Println("Detecting current package manager")

    	// List of common package managers
	packageManagers := []string{
		"apt-get", "apt", "yum", "dnf", "zypper", "pacman", "brew", "apk", "pkg",
	}

	for _, pm := range packageManagers {
		cmd := exec.Command("which", pm)
		err := cmd.Run()
		if err == nil {
			return pm, nil
		}
	}

	return "", fmt.Errorf("no supported package manager found")
}

func pkgManager(action, software string) error {
	pm, err := detectPackageManager()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	var queryCmd string
	
	switch pm {
	case "apt-get", "apt":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			queryCmd = fmt.Sprintf("dpkg -l | grep ^ii | awk '{print $2}' | grep ^%s$", software)
			cmd = exec.Command("sh", "-c", queryCmd)
		}
	case "yum":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("rpm", "-q", software)
		}
	case "dnf":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("dnf", "list", "installed", software)
		}
	case "zypper":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("zypper", "se", "--installed-only", software)
		}
	case "pacman":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "-S", "--noconfirm", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "-R", "--noconfirm", software)
		} else if action == "check" {
			cmd = exec.Command("pacman", "-Q", software)
		}
	case "brew":
		if action == "install" {
			cmd = exec.Command("brew", "install", software)
		} else if action == "remove" {
			cmd = exec.Command("brew", "uninstall", software)
		} else if action == "check" {
			cmd = exec.Command("brew", "list", software)
		}
	case "apk":
		if action == "install" {
			cmd = exec.Command("apk", "add", software)
		} else if action == "remove" {
			cmd = exec.Command("apk", "del", software)
		} else if action == "check" {
			cmd = exec.Command("apk", "info", software)
		}
	case "pkg":
		if action == "install" {
			cmd = exec.Command("pkg", "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("pkg", "delete", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("pkg", "info", software)
		}
	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	if cmd == nil {
		return fmt.Errorf("no command found for package manager: %s", pm)
	}

	if action == "check" {
		_, err := cmd.CombinedOutput()
		if err != nil {
			return err
		} else {
			log.Println("Running package manager %s", cmd)
			return nil
		}
	}
	return err;
}

