package cli

import (
	"bufio"
	"fmt"
	"os"
	"igloo/tui"
	//"igloo/Test"
	"strings"
)

func Main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		arguments := strings.Split(strings.TrimSpace(input), " ")
		switch arguments[0] {
		case "test":
		//test.Main()
		case "tui":
			tui.UI()
		default:
			fmt.Println(arguments)
		}
	}

}
