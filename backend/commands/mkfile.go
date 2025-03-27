package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/stores"
	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/structures"
	utils "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/utils"
)

// MKFILE representa el comando mkfile con sus parámetros
type MKFILE struct {
	path string // Ruta absoluta del archivo
	r    bool   // Crear carpetas padre recursivamente
	size int    // Tamaño en bytes (default 0)
	cont string // Contenido del archivo
}

// ParseMkfile parsea los tokens del comando mkfile
func ParseMkfile(tokens []string) (string, error) {
	cmd := &MKFILE{}

	// Unir tokens para manejar espacios en parámetros como -cont
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=(?:"[^"]+"|[^\s]+)|-size=[0-9]+|-cont=(?:"[^"]+"|[^\s]+)|-r`)
	matches := re.FindAllString(args, -1)

	// Validar que todos los tokens sean parámetros válidos
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

		switch key {
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato inválido para -path: %s", match)
			}
			value := strings.Trim(kv[1], "\"")
			if !strings.HasPrefix(value, "/") {
				return "", errors.New("la ruta debe ser absoluta (comenzar con /)")
			}
			cmd.path = value
		case "-size":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato inválido para -size: %s", match)
			}
			size, err := utils.StringToInt(kv[1])
			if err != nil || size < 0 {
				return "", fmt.Errorf("tamaño inválido: %s", kv[1])
			}
			cmd.size = size
		case "-cont":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato inválido para -cont: %s", match)
			}
			value := strings.Trim(kv[1], "\"")
			cmd.cont = value
		case "-r":
			cmd.r = true
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verificar parámetro obligatorio
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecutar el comando
	err := commandMkfile(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKFILE: Archivo %s creado correctamente", cmd.path), nil
}

// commandMkfile implementa la lógica del comando mkfile
func commandMkfile(mkfile *MKFILE) error {
	// Verificar sesión activa
	if stores.CurrentSession.ID == "" {
		return errors.New("debe iniciar sesión primero")
	}

	// Obtener la partición montada
	sb, mountedPartition, diskPath, err := stores.GetMountedPartitionSuperblock(stores.CurrentSession.ID)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Separar directorios padres y nombre del archivo
	parentDirs, fileName := utils.GetParentDirectories(mkfile.path)
	fmt.Printf("Creando archivo: %s\n", mkfile.path)
	fmt.Printf("Directorios padres: %v\n", parentDirs)
	fmt.Printf("Nombre del archivo: %s\n", fileName)

	// Manejar directorios padres
	if len(parentDirs) > 0 {
		if mkfile.r {
			// Crear carpetas recursivamente
			err = createParentFolders(sb, diskPath, parentDirs)
			if err != nil {
				return fmt.Errorf("error al crear directorios padres: %w", err)
			}
			// Serializar superbloque tras crear carpetas
			err = sb.Serialize(diskPath, int64(mountedPartition.Part_start))
			if err != nil {
				return fmt.Errorf("error al serializar superbloque tras crear carpetas: %w", err)
			}
		} else {
			// Verificar si existen sin -r
			if !checkParentExists(sb, diskPath, parentDirs) {
				return fmt.Errorf("el directorio padre %s no existe (use -r para crearlo)", strings.Join(parentDirs, "/"))
			}
		}
	}

	// Determinar contenido final
	finalContent := ""
	if mkfile.cont != "" {
		finalContent = mkfile.cont
	} else if mkfile.size > 0 {
		finalContent = strings.Repeat("0", mkfile.size)
	}

	// Crear el archivo
	err = createFile(sb, diskPath, parentDirs, fileName, finalContent)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Imprimir estado (para depuración)
	sb.PrintInodes(diskPath)
	sb.PrintBlocks(diskPath)

	// Serializar superbloque
	err = sb.Serialize(diskPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar superbloque: %w", err)
	}

	return nil
}

// checkParentExists verifica si los directorios padres existen
func checkParentExists(sb *structures.SuperBlock, diskPath string, parentDirs []string) bool {
	currentInode := int32(0) // Raíz
	for _, dir := range parentDirs {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
		if err != nil || inode.I_type[0] != '0' { // Debe ser carpeta
			return false
		}
		found := false
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			block := &structures.FolderBlock{}
			err = block.Deserialize(diskPath, int64(sb.S_block_start+blockIndex*sb.S_block_size))
			if err != nil {
				return false
			}
			for _, content := range block.B_content {
				name := strings.Trim(string(content.B_name[:]), "\x00")
				if name == dir && content.B_inodo != -1 {
					currentInode = content.B_inodo
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// createParentFolders crea los directorios padres recursivamente
func createParentFolders(sb *structures.SuperBlock, diskPath string, parentDirs []string) error {
	currentInode := int32(0) // Raíz
	for _, dir := range parentDirs {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
		if err != nil {
			return err
		}

		// Verificar si la carpeta ya existe
		found := false
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			block := &structures.FolderBlock{}
			err = block.Deserialize(diskPath, int64(sb.S_block_start+blockIndex*sb.S_block_size))
			if err != nil {
				return err
			}
			for _, content := range block.B_content {
				name := strings.Trim(string(content.B_name[:]), "\x00")
				if name == dir && content.B_inodo != -1 {
					currentInode = content.B_inodo
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		// Si no existe, crear la carpeta
		if !found {
			uid, err := strconv.Atoi(stores.CurrentSession.UID)
			if err != nil {
				return fmt.Errorf("error convirtiendo UID: %v", err)
			}
			gid, err := strconv.Atoi(stores.CurrentSession.GID)
			if err != nil {
				return fmt.Errorf("error convirtiendo GID: %v", err)
			}

			// Crear nuevo inodo para la carpeta
			newInode := &structures.Inode{
				I_uid:   int32(uid),
				I_gid:   int32(gid),
				I_size:  0,
				I_atime: float32(time.Now().Unix()),
				I_ctime: float32(time.Now().Unix()),
				I_mtime: float32(time.Now().Unix()),
				I_type:  [1]byte{'0'},           // Carpeta
				I_perm:  [3]byte{'6', '6', '4'}, // Permisos 664
			}
			newInodeIndex := sb.S_inodes_count

			// Crear bloque inicial para la carpeta (con . y ..)
			newBlock := &structures.FolderBlock{
				B_content: [4]structures.FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: newInodeIndex},
					{B_name: [12]byte{'.', '.'}, B_inodo: currentInode},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}
			newBlockIndex := sb.S_blocks_count
			newInode.I_block[0] = newBlockIndex

			// Buscar espacio en el inodo padre o asignar un nuevo bloque
			var parentBlockIndex int32 = -1
			var parentBlock *structures.FolderBlock
			for j, bIndex := range inode.I_block {
				if bIndex != -1 {
					parentBlock = &structures.FolderBlock{}
					err = parentBlock.Deserialize(diskPath, int64(sb.S_block_start+bIndex*sb.S_block_size))
					if err != nil {
						return err
					}
					for k := 0; k < len(parentBlock.B_content); k++ {
						if parentBlock.B_content[k].B_inodo == -1 || strings.Trim(string(parentBlock.B_content[k].B_name[:]), "\x00") == "" {
							parentBlockIndex = bIndex
							parentBlock.B_content[k].B_inodo = newInodeIndex
							copy(parentBlock.B_content[k].B_name[:], dir)
							break
						}
					}
					if parentBlockIndex != -1 {
						break
					}
				} else if j < 12 { // Usar bloques directos
					parentBlock = &structures.FolderBlock{
						B_content: [4]structures.FolderContent{
							{B_name: [12]byte{}, B_inodo: newInodeIndex},
							{B_name: [12]byte{'-'}, B_inodo: -1},
							{B_name: [12]byte{'-'}, B_inodo: -1},
							{B_name: [12]byte{'-'}, B_inodo: -1},
						},
					}
					inode.I_block[j] = sb.S_blocks_count
					parentBlockIndex = sb.S_blocks_count
					copy(parentBlock.B_content[0].B_name[:], dir)

					err = parentBlock.Serialize(diskPath, int64(sb.S_block_start+parentBlockIndex*sb.S_block_size))
					if err != nil {
						return err
					}
					err = sb.UpdateBitmapBlock(diskPath)
					if err != nil {
						return err
					}
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					break
				}
			}
			if parentBlockIndex == -1 {
				return errors.New("no hay espacio en el directorio padre para crear " + dir)
			}

			// Serializar nuevo bloque
			err = newBlock.Serialize(diskPath, int64(sb.S_block_start+newBlockIndex*sb.S_block_size))
			if err != nil {
				return err
			}
			err = sb.UpdateBitmapBlock(diskPath)
			if err != nil {
				return err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--

			// Serializar nuevo inodo
			err = newInode.Serialize(diskPath, int64(sb.S_inode_start+newInodeIndex*sb.S_inode_size))
			if err != nil {
				return err
			}
			err = sb.UpdateBitmapInode(diskPath)
			if err != nil {
				return err
			}
			sb.S_inodes_count++
			sb.S_free_inodes_count--

			// Actualizar bloque padre
			err = parentBlock.Serialize(diskPath, int64(sb.S_block_start+parentBlockIndex*sb.S_block_size))
			if err != nil {
				return err
			}

			// Actualizar inodo padre
			err = inode.Serialize(diskPath, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
			if err != nil {
				return err
			}

			currentInode = newInodeIndex
		}
	}
	return nil
}

// createFile crea un archivo en el sistema de archivos
func createFile(sb *structures.SuperBlock, diskPath string, parentDirs []string, fileName string, content string) error {
	currentInode := int32(0)
	for _, dir := range parentDirs {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
		if err != nil {
			return err
		}
		found := false
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			block := &structures.FolderBlock{}
			err = block.Deserialize(diskPath, int64(sb.S_block_start+blockIndex*sb.S_block_size))
			if err != nil {
				return err
			}
			for _, content := range block.B_content {
				name := strings.Trim(string(content.B_name[:]), "\x00")
				if name == dir && content.B_inodo != -1 {
					currentInode = content.B_inodo
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("directorio %s no encontrado", dir)
		}
	}

	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
	if err != nil {
		return err
	}

	var targetBlockIndex int32 = -1
	var blockToUpdate *structures.FolderBlock
	var contentIndex int
	for i, blockIndex := range inode.I_block {
		if blockIndex != -1 {
			block := &structures.FolderBlock{}
			err = block.Deserialize(diskPath, int64(sb.S_block_start+blockIndex*sb.S_block_size))
			if err != nil {
				return err
			}
			for j := 0; j < len(block.B_content); j++ {
				if block.B_content[j].B_inodo == -1 || strings.Trim(string(block.B_content[j].B_name[:]), "\x00") == "" {
					targetBlockIndex = blockIndex
					blockToUpdate = block
					contentIndex = j
					break
				}
			}
			if targetBlockIndex != -1 {
				break
			}
		} else if i < 12 {
			newBlock := &structures.FolderBlock{}
			inode.I_block[i] = sb.S_blocks_count
			targetBlockIndex = sb.S_blocks_count
			blockToUpdate = newBlock
			contentIndex = 0

			err = newBlock.Serialize(diskPath, int64(sb.S_block_start+targetBlockIndex*sb.S_block_size))
			if err != nil {
				return err
			}
			err = sb.UpdateBitmapBlock(diskPath)
			if err != nil {
				return err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			err = inode.Serialize(diskPath, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
			if err != nil {
				return err
			}
			break
		}
	}
	if targetBlockIndex == -1 {
		return errors.New("no hay espacio en el directorio padre para crear " + fileName)
	}

	uid, err := strconv.Atoi(stores.CurrentSession.UID)
	if err != nil {
		return fmt.Errorf("error convirtiendo UID: %v", err)
	}
	gid, err := strconv.Atoi(stores.CurrentSession.GID)
	if err != nil {
		return fmt.Errorf("error convirtiendo GID: %v", err)
	}
	fileInode := &structures.Inode{
		I_uid:   int32(uid),
		I_gid:   int32(gid),
		I_size:  int32(len(content)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'6', '6', '4'},
	}
	fileInodeIndex := sb.S_inodes_count

	blockToUpdate.B_content[contentIndex].B_inodo = fileInodeIndex
	copy(blockToUpdate.B_content[contentIndex].B_name[:], fileName)
	err = blockToUpdate.Serialize(diskPath, int64(sb.S_block_start+targetBlockIndex*sb.S_block_size))
	if err != nil {
		return err
	}

	// Si no hay contenido, aún asignamos un bloque vacío
	if content == "" {
		fileBlock := &structures.FileBlock{
			B_content: [64]byte{}, // Bloque vacío
		}
		fileInode.I_block[0] = sb.S_blocks_count
		err = fileBlock.Serialize(diskPath, int64(sb.S_block_start+sb.S_blocks_count*sb.S_block_size))
		if err != nil {
			return err
		}
		err = sb.UpdateBitmapBlock(diskPath)
		if err != nil {
			return err
		}
		sb.S_blocks_count++
		sb.S_free_blocks_count--
	} else if content != "" {
		chunks := utils.SplitStringIntoChunks(content)
		if len(chunks) > 12 {
			return fmt.Errorf("contenido demasiado grande, máximo 12 bloques directos")
		}
		for i, chunk := range chunks {
			fileBlock := &structures.FileBlock{}
			copy(fileBlock.B_content[:], chunk)
			fileInode.I_block[i] = sb.S_blocks_count
			err = fileBlock.Serialize(diskPath, int64(sb.S_block_start+sb.S_blocks_count*sb.S_block_size))
			if err != nil {
				return err
			}
			err = sb.UpdateBitmapBlock(diskPath)
			if err != nil {
				return err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--
		}
	}

	err = fileInode.Serialize(diskPath, int64(sb.S_inode_start+fileInodeIndex*sb.S_inode_size))
	if err != nil {
		return err
	}
	err = sb.UpdateBitmapInode(diskPath)
	if err != nil {
		return err
	}
	sb.S_inodes_count++
	sb.S_free_inodes_count--

	return nil
}
