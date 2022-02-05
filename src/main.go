package main

import (
	"fmt"
	"krypt/core"
	"log"
	"os"

	"github.com/tristanisham/colors"
)

func main() {
	args := os.Args[1:]
	lock := core.New()
	if len(args) < 2 {
		help_text("Insufficent number of arguments.\n")
		os.Exit(2)
	}

	switch args[0] {
	case "u", "unlock":
		lock.Action = core.Unlock
	case "l", "lock":
		lock.Action = core.Lock
	default:
		help_text("Help:.\n")
		os.Exit(2)
	}
	for c, v := range args {
		if v == "-i" && len(args) >= c+1 {
			lock.Input = args[c+1]
		} else if v == "-o" && len(args) >= c+1 {
			lock.Output = args[c+1]
			// } else if v == "-k" && len(args) >= c+1 {
			// 	lock.Key = args[c+1]
		}
	}

	// if len(lock.Key) == 0 {
	// 	log.Fatal("Please provide a Key")
	// }

	if err := lock.Start(); err != nil {
		log.Fatal(err)
	}
}

func help_text(message string) {
	fmt.Println(
		message,
		"\t", colors.As("u", colors.Bold), "|", colors.As("unlock", colors.Bold), " => Sets krypt mode to delock\n",
		"\t", colors.As("l", colors.Bold), "|", colors.As("lock", colors.Bold), " => Sets krypt mode to enlock\n",
		"\t", colors.As("-i", colors.Bold), "| => Points to the file krypt should action on",
	)
}
