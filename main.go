package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"os"
)

func main() {
	Args := os.Args

	// If there are two arguments and the first is "-command" then
	// go into command mode
	if len(Args) > 2 && Args[1] == "-command" {
		fmt.Print(command_mode(Args[2], Args[3:]))
		return
	}

	log.Println("Lambda started")

	lambda.StartHandler(Handler{})
}
