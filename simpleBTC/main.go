package main

import (
	"bufio"
	"fmt"
	"github.com/szh7654/simpleBTC/BLC"
	"log"
	"os"
)

func main() {
	cli := BLC.CLI{}
	cli.Run()

}