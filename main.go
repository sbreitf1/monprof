package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/manifoldco/promptui"
	"github.com/sbreitf1/exec"
	"github.com/sebidude/configparser"
)

var (
	argConfigFile = kingpin.Flag("config", "Config file path.").Short('c').String()
	argSelect     = kingpin.Flag("select", "Select profile manually").Short('s').Bool()
	appConfig     config
)

func main() {
	kingpin.Parse()

	if len(*argConfigFile) == 0 {
		fmt.Println("Use default config file monprof.conf")
		*argConfigFile = "monprof.conf"
	}
	if err := configparser.ParseYaml(*argConfigFile, &appConfig); err != nil {
		log.Fatalf("cannot parse configfile %q: %s", *argConfigFile, err.Error())
	}

	if len(appConfig.Profiles) == 0 {
		log.Fatalln("No profiles configured")
	}

	fmt.Println("Found " + strconv.Itoa(len(appConfig.Profiles)) + " profiles:")
	for _, profile := range appConfig.Profiles {
		fmt.Println(" -> " + profile.Name)
	}

	var selectedProfile *profile
	if *argSelect {
		promptList := make([]string, len(appConfig.Profiles)+1)
		promptList[0] = "(abort)"
		for i := range appConfig.Profiles {
			promptList[i+1] = appConfig.Profiles[i].Name
		}
		ui := promptui.Select{Label: "Select monitor profile", Items: promptList}
		index, _, err := ui.Run()
		if err != nil {
			log.Fatalln("Failed to interactively select profile:", err.Error())
		}
		if index == 0 {
			fmt.Println("no profile selected")
			os.Exit(0)
		}
		selectedProfile = &appConfig.Profiles[index-1]

	} else {
		output, err := execCmd("xrandr")
		if err != nil {
			log.Fatalf("an error occured while running xrandr: %s", err.Error())
		}
		outStr := string(output)

		for _, profile := range appConfig.Profiles {
			if profile.Conditions == nil || len(profile.Conditions) == 0 {
				// configuration to be selected manually
				continue
			}

			skip := false
			for _, condition := range profile.Conditions {
				if !strings.Contains(outStr, condition.Monitor+" connected") {
					skip = true
					continue
				}
			}

			if skip {
				// at least one condition is not fulfilled -> skip this profile
				continue
			}

			selectedProfile = &profile
			break
		}
	}

	if selectedProfile != nil {
		fmt.Println("Select profile " + selectedProfile.Name)
		for _, cmd := range selectedProfile.Commands {
			if _, err := execCmd(cmd); err != nil {
				log.Fatalf("failed to execute %q: %s", cmd, err.Error())
			}
		}
		fmt.Println(" -> Okay")
	} else {
		log.Fatalln("No profile found for the current monitor configuration")
	}
}

func execCmd(cmd string) (string, error) {
	result, code, err := exec.RunLine(cmd)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("Command %q exited with code %d", cmd, code)
	}
	return result, nil
}
