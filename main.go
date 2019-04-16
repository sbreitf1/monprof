package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sebidude/configparser"
)

var (
	argConfigFile = kingpin.Flag("config", "Config file path.").Short('c').String()
	config        Config
)

func main() {
	kingpin.Parse()

	if err := configparser.ParseYaml(*argConfigFile, &config); err != nil {
		log.Fatalf("cannot parse configfile %q: %s", *argConfigFile, err.Error())
	}

	if len(config.Profiles) == 0 {
		log.Fatalln("No profiles configured")
	}

	fmt.Println("Found " + strconv.Itoa(len(config.Profiles)) + " profiles:")
	for _, profile := range config.Profiles {
		fmt.Println(" -> " + profile.Name)
	}

	cmd := exec.Command("xrandr")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("an error occured while running xrandr: %s", err.Error())
	}
	outStr := string(output)

	var selectedProfile *Profile
	for _, profile := range config.Profiles {
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

	if selectedProfile != nil {
		fmt.Println("Select profile " + selectedProfile.Name)
		for _, cmd := range selectedProfile.Commands {
			if err := execCmd(cmd); err != nil {
				log.Fatalf("failed to execute %q: %s", cmd, err.Error())
			}
		}
		fmt.Println(" -> Okay")
	} else {
		log.Fatalln("No profile found for the current monitor configuration")
	}
}

func execCmd(cmd string) error {
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	parts := r.FindAllString(cmd, -1)
	args := make([]string, len(parts))
	for i, part := range parts {
		args[i] = part
		if strings.HasPrefix(args[i], `"`) {
			args[i] = args[i][1:]
		}
		if strings.HasSuffix(args[i], `"`) {
			args[i] = args[i][:len(args[i])-1]
		}
	}

	c := exec.Command(args[0], args[1:]...)
	output, err := c.Output()
	if err != nil {
		log.Println(string(output))
		return err
	}

	return nil
}
