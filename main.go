package main

import (
	"bufio" //package that contains functions for reading input from the user
	"fmt"   //package that contains functions for printing formatted output and scanning input
	"os"    //package that contains functions for interacting with the operating system

	analyzer "github.com/MarceJua/MIA_1S2025_P1_202010367/analyzer" //import the package that contains the functions to analyze the input from the user
)

func main() {
	scanner := bufio.NewScanner(os.Stdin) //create a new scanner object to read input from the user

	for {
		fmt.Print(">>> ") //print the prompt

		if !scanner.Scan() { //scan the input from the user
			break //if there is an error, break the loop
		}

		input := scanner.Text() //get the input from the user

		_, err := analyzer.Analyzer(input) //scan the input from the user
		if err != nil {
			fmt.Println("Error: ", err) //print the error message
			continue
		}

		//Aqui se podria imprimir el comando analizado
		//fmt.Println("Parsed Command: %+v\n", cmd)

	}

	if err := scanner.Err(); err != nil { //check if there is an error
		fmt.Fprintln(os.Stderr, "reading standard input:", err) //print the error message
		//fmt.Println("Error al leer:", err)
	}
}
