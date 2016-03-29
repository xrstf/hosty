package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cmdHash() {
	fmt.Print("   Password: ")

	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err == nil {
		password = strings.TrimSpace(password)
		fmt.Print("bcrypt hash: " + string(hashBcrypt(password)) + "\n")
	}
}
