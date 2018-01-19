package setup

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// PerformSetup will start setting up HouseKeeper for the user.
func PerformSetup() {
	fmt.Println("Welcome to HouseKeeper, performing account setup...")

	if getYes("Setup HouseKeeper in AWS?", true) {
		err := awsSetup()
		if err != nil {
			fmt.Printf("AWS setup failed: %s\n", err)
			os.Exit(1)
		}
	}
	fmt.Println(`
SUCCESS

Nothing else to setup, all done! :)
`)
}

func getYes(prompt string, yesDefault bool) bool {
	reader := bufio.NewReader(os.Stdin)
	if yesDefault {
		prompt = fmt.Sprintf("%s (Y/n): ", prompt)
	} else {
		prompt = fmt.Sprintf("%s (y/N): ", prompt)
	}
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln(err)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return yesDefault
	}
	return strings.Contains(input, "y")
}
