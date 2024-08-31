package util

import "log"

func CheckIfInstalled(serviceName string) bool {
	// Check if the software is installed.
	err := pkgManager("check", serviceName)
	if err == nil {
		return true
	} else {
		return false
	}
}

func CheckAndInstallSoftware(packages []string) {
	allowedPackages := map[string]struct{}{
        "nginx":           {},
        "mariadb-server":  {},
        "php8.1-fpm":      {},
        "cron":            {},
    }
		for _, pkg := range packages {
			if _, ok := allowedPackages[pkg]; ok {
				if !CheckIfInstalled(pkg) {
					err := pkgManager("install", pkg)
					log.Printf("installing: %s \n", pkg)
					if err != nil {
						log.Printf("Error: %s", err)
					}
					} else {
					log.Printf("%s is already installed.\n", pkg)
				}
			}  else {
				log.Printf("%s is not in the allowed list.\n", pkg)
			}
		}
}

func CheckIfStackReady () (bool) {
	if CheckIfInstalled("nginx") && CheckIfInstalled("mariadb-server") && CheckIfInstalled("php8.1-fpm") && CheckIfInstalled("cron") {
		return true
	} else {
		return false
	}
}