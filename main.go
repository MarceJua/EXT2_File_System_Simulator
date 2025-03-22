package main

import (
	"bufio" //package that contains functions for reading input from the user
	"fmt"   //package that contains functions for printing formatted output and scanning input
	"os"    //package that contains functions for interacting with the operating system

	analyzer "github.com/MarceJua/MIA_1S2025_P1_202010367/analyzer" //import the package that contains the functions to analyze the input from the user
)

func main() {
	// Crea un nuevo escáner que lee desde la entrada estándar (teclado)
	scanner := bufio.NewScanner(os.Stdin)

	// Bucle infinito para leer comandos del usuario
	for {
		fmt.Print(">>> ") // Imprime el prompt para el usuario

		// Lee la siguiente línea de entrada del usuario
		if !scanner.Scan() {
			break // Si no hay más líneas para leer, rompe el bucle
		}

		// Obtiene el texto ingresado por el usuario
		input := scanner.Text()

		// Llama a la función Analyzer del paquete analyzer para analizar el comando ingresado
		result, err := analyzer.Analyzer(input)
		if err != nil {
			// Si hay un error al analizar el comando, imprime el error y continúa con el siguiente comando
			fmt.Println("Error:", err)
			continue
		}
		// Imprimir el resultado si no hay error
		if result != nil {
			fmt.Println(result)
		}
	}

	// Verifica si hubo algún error al leer la entrada
	if err := scanner.Err(); err != nil {
		// Si hubo un error al leer la entrada, lo imprime
		fmt.Println("Error al leer:", err)
	}
}
