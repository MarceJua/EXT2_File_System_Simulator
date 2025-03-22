package analyzer

import (
	"errors"  // Importa el paquete "errors" para manejar errores
	"fmt"     // Importa el paquete "fmt" para formatear e imprimir texto
	"strings" // Importa el paquete "strings" para manipulación de cadenas

	commands "github.com/MarceJua/MIA_1S2025_P1_202010367/commands" // Importa el paquete "commands" que contiene las funciones para analizar comandos
)

// splitCommand divide la entrada respetando cadenas entre comillas
func splitCommand(input string) []string {
	var tokens []string
	var currentToken strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		switch char {
		case '"':
			inQuotes = !inQuotes
			currentToken.WriteByte(char)
		case ' ':
			if inQuotes {
				currentToken.WriteByte(char)
			} else if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		default:
			currentToken.WriteByte(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

// Analyzer analiza el comando de entrada y ejecuta la acción correspondiente
func Analyzer(input string) (interface{}, error) {
	// Divide la entrada en tokens usando espacios en blanco como delimitadores
	//tokens := strings.Fields(input)
	tokens := splitCommand(input)

	// Si no se proporcionó ningún comando, devuelve un error
	if len(tokens) == 0 {
		return nil, errors.New("no se proporcionó ningún comando")
	}

	// Switch para manejar diferentes comandos
	switch tokens[0] {
	case "mkdisk":
		// Llama a la función ParseMkdisk del paquete commands con los argumentos restantes
		return commands.ParseMkdisk(tokens[1:])
	case "rmdisk":
		// Llama a la función ParseRmdisk del paquete commands con los argumentos restantes
		return commands.ParseRmdisk(tokens[1:])
	case "fdisk":
		// Llama a la función CommandFdisk del paquete commands con los argumentos restantes
		return commands.ParseFdisk(tokens[1:])
	case "mount":
		// Llama a la función CommandMount del paquete commands con los argumentos restantes
		return commands.ParseMount(tokens[1:])
	case "mounted":
		// Llama a la función CommandMounted del paquete commands con los argumentos restantes
		return commands.ParseMounted(tokens[1:])
	case "mkfs":
		// Llama a la función CommandMkfs del paquete commands con los argumentos restantes
		return commands.ParseMkfs(tokens[1:])
	case "rep":
		// Llama a la función CommandRep del paquete commands con los argumentos restantes
		return commands.ParseRep(tokens[1:])
	case "mkdir":
		// Llama a la función CommandMkdir del paquete commands con los argumentos restantes
		return commands.ParseMkdir(tokens[1:])
	case "login":
		// Llama a la función CommandLogin del paquete commands con los argumentos restantes
		return commands.ParseLogin(tokens[1:])
	case "logout":
		// Llama a la función CommandLogout del paquete commands con los argumentos restantes
		return commands.ParseLogout(tokens[1:])
	case "mkfile":
		// Llama a la función CommandMkfile del paquete commands con los argumentos restantes
		return commands.ParseMkfile(tokens[1:])
	case "cat":
		// Llama a la función CommandCat del paquete commands con los argumentos restantes
		return commands.ParseCat(tokens[1:])
	default:
		// Si el comando no es reconocido, devuelve un error
		return nil, fmt.Errorf("comando desconocido: %s", tokens[0])
	}
}
