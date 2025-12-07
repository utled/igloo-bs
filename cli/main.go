package cli

import (
	"bufio"
	"fmt"
	"os"
	"snafu/indexing"
	"snafu/setup"
	"snafu/test"
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
			test.Main()
		case "setup":
			err := setup.Main()
			if err != nil {
				fmt.Println(err)
			}
		case "fullscan":
			indexing.Main()
		default:
			fmt.Println(arguments)
		}
	}

}
