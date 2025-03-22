package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/stores"
	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/structures"
)

// LOGIN estructura que representa el comando login con sus parámetros
type LOGIN struct {
	user string // Nombre del usuario
	pass string // Contraseña
	id   string // ID de la partición
}

// ParseLogin parsea los tokens del comando login
func ParseLogin(tokens []string) (string, error) {
	cmd := &LOGIN{}

	// Unir tokens en una sola cadena
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user=[^\s]+|-pass=[^\s]+|-id=[^\s]+`)
	matches := re.FindAllString(args, -1)

	// Verificar tokens válidos
	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	// Parsear parámetros
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		value := kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-user":
			cmd.user = value
		case "-pass":
			cmd.pass = value
		case "-id":
			cmd.id = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verificar parámetros requeridos
	if cmd.user == "" || cmd.pass == "" || cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -user, -pass, -id")
	}

	// Ejecutar el comando
	err := commandLogin(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("LOGIN: Sesión iniciada como %s en %s", cmd.user, cmd.id), nil
}

// commandLogin implementa la lógica del comando login
func commandLogin(login *LOGIN) error {
	// Verificar si ya hay una sesión activa
	if stores.CurrentSession.ID != "" {
		return errors.New("ya hay una sesión activa, cierre la sesión actual primero")
	}

	// Obtener la partición montada
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(login.id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Buscar el inodo de users.txt (asumimos que es el inodo 1)
	usersInode := &structures.Inode{}
	err = usersInode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+partitionSuperblock.S_inode_size)) // Inodo 1
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %w", err)
	}

	// Verificar que sea un archivo
	if usersInode.I_type[0] != '1' {
		return errors.New("users.txt no es un archivo válido")
	}

	// Leer el bloque de datos de users.txt
	blockIndex := usersInode.I_block[0]
	if blockIndex == -1 {
		return errors.New("no se encontró contenido en users.txt")
	}
	fmt.Printf("I_block[0] de users.txt: %d\n", blockIndex)

	fileBlock := &structures.FileBlock{}
	err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+blockIndex*partitionSuperblock.S_block_size))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de users.txt: %w", err)
	}

	// Obtener el contenido como string
	content := strings.Trim(string(fileBlock.B_content[:]), "\x00")
	fmt.Printf("Contenido de users.txt: %s\n", content)

	// Procesar las líneas de users.txt
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		fmt.Printf("Línea procesada: %v\n", parts)

		// Verificar si es un usuario (formato: ID,U,username,password)
		if len(parts) == 4 && parts[1] == "U" {
			username := parts[2]
			password := parts[3]
			if username == login.user && password == login.pass {
				// Guardar la sesión
				stores.CurrentSession = stores.Session{
					ID:       login.id,
					Username: login.user,
					UID:      parts[0], // ID del usuario
					GID:      parts[0], // Usamos el mismo ID para el grupo por simplicidad
				}
				return nil
			}
		}
	}

	return fmt.Errorf("usuario o contraseña incorrectos")
}
