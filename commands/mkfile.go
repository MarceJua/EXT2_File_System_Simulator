package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/stores"
	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/structures"
	utils "github.com/MarceJua/MIA_1S2025_P1_202010367/utils"
)

// MKFILE representa el comando mkfile con sus parámetros
type MKFILE struct {
	path string // Ruta del archivo
	p    bool   // Opción -p para crear directorios padres
	size int    // Tamaño en bytes
	cont string // Contenido del archivo
}

// ParseMkfile parsea los tokens del comando mkfile
func ParseMkfile(tokens []string) (string, error) {
	cmd := &MKFILE{}

	// Unir tokens para manejar parámetros con espacios
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=(?:"[^"]+"|[^\s]+)|-p|-size=[0-9]+|-cont=(?:"[^"]+"|[^\s]+)`)
	matches := re.FindAllString(args, -1)

	// Validar tokens
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
			cmd.path = value
		case "-p":
			cmd.p = true
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
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verificar parámetro requerido
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
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(stores.CurrentSession.ID)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Separar directorios padres y nombre del archivo
	parentDirs, fileName := utils.GetParentDirectories(mkfile.path)
	fmt.Println("\nCreando archivo:", mkfile.path)
	fmt.Println("Directorios padres:", parentDirs)
	fmt.Println("Nombre del archivo:", fileName)

	// Crear directorios padres si es necesario
	if len(parentDirs) > 0 {
		if mkfile.p {
			for i := 0; i < len(parentDirs); i++ {
				currentPath := parentDirs[:i+1] // e.g., ["/folder"]
				if !checkParentExists(partitionSuperblock, partitionPath, currentPath) {
					var parentPath []string
					var folderName string
					if i == 0 {
						parentPath = []string{} // Raíz
						folderName = parentDirs[0]
					} else {
						parentPath = parentDirs[:i] // Camino hasta el padre
						folderName = parentDirs[i]
					}
					err = partitionSuperblock.CreateFolder(partitionPath, parentPath, folderName)
					if err != nil {
						return fmt.Errorf("error al crear directorio padre %s: %w", folderName, err)
					}
					// Forzar serialización del superbloque para reflejar cambios
					err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
					if err != nil {
						return fmt.Errorf("error al serializar superbloque tras crear %s: %w", folderName, err)
					}
				}
			}
		} else {
			if !checkParentExists(partitionSuperblock, partitionPath, parentDirs) {
				return fmt.Errorf("el directorio padre %s no existe (use -p para crearlo)", strings.Join(parentDirs, "/"))
			}
		}
	}

	// Crear el archivo
	err = createFile(partitionSuperblock, partitionPath, parentDirs, fileName, mkfile.size, mkfile.cont)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Imprimir inodos y bloques
	partitionSuperblock.PrintInodes(partitionPath)
	partitionSuperblock.PrintBlocks(partitionPath)

	// Serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}

// checkParentExists verifica si los directorios padres existen
func checkParentExists(sb *structures.SuperBlock, path string, parentDirs []string) bool {
	currentInode := int32(0) // Empezar desde la raíz
	for _, dir := range parentDirs {
		inode := &structures.Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
		if err != nil || inode.I_type[0] != '0' {
			return false
		}
		found := false
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			block := &structures.FolderBlock{}
			err = block.Deserialize(path, int64(sb.S_block_start+blockIndex*sb.S_block_size))
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

// createFile crea un archivo en el sistema de archivos
func createFile(sb *structures.SuperBlock, path string, parentDirs []string, fileName string, size int, content string) error {
	// Encontrar el inodo padre
	currentInode := int32(0) // Empezar desde la raíz
	for _, dir := range parentDirs {
		inode := &structures.Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
		if err != nil {
			return err
		}
		found := false
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			block := &structures.FolderBlock{}
			err = block.Deserialize(path, int64(sb.S_block_start+blockIndex*sb.S_block_size))
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

	// Crear el archivo en el inodo padre
	inode := &structures.Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
	if err != nil {
		return err
	}

	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 && i > 0 { // Crear nuevo bloque si es necesario
			newBlock := &structures.FolderBlock{}
			inode.I_block[i] = sb.S_blocks_count
			err = newBlock.Serialize(path, int64(sb.S_first_blo))
			if err != nil {
				return err
			}
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size
			err = inode.Serialize(path, int64(sb.S_inode_start+currentInode*sb.S_inode_size))
			if err != nil {
				return err
			}
			blockIndex = inode.I_block[i]
		}
		if blockIndex != -1 {
			block := &structures.FolderBlock{}
			err = block.Deserialize(path, int64(sb.S_block_start+blockIndex*sb.S_block_size))
			if err != nil {
				return err
			}
			for j := 2; j < len(block.B_content); j++ {
				if block.B_content[j].B_inodo == -1 {
					// Crear nuevo inodo para el archivo
					fileInode := &structures.Inode{
						I_uid:   1,
						I_gid:   1,
						I_size:  int32(size),
						I_atime: float32(time.Now().Unix()),
						I_ctime: float32(time.Now().Unix()),
						I_mtime: float32(time.Now().Unix()),
						I_type:  [1]byte{'1'},
						I_perm:  [3]byte{'6', '6', '4'},
					}
					fileInodeIndex := sb.S_inodes_count
					block.B_content[j] = structures.FolderContent{
						B_name:  [12]byte{},
						B_inodo: fileInodeIndex,
					}
					copy(block.B_content[j].B_name[:], fileName)
					err = block.Serialize(path, int64(sb.S_block_start+blockIndex*sb.S_block_size))
					if err != nil {
						return err
					}

					// Manejar contenido
					finalContent := content
					if content == "" && size > 0 {
						finalContent = strings.Repeat("0", size)
					}
					if finalContent != "" {
						chunks := utils.SplitStringIntoChunks(finalContent)
						if len(chunks) > 15 {
							return fmt.Errorf("contenido demasiado grande, máximo 15 bloques")
						}
						for i, chunk := range chunks {
							fileBlock := &structures.FileBlock{}
							copy(fileBlock.B_content[:], chunk)
							fileInode.I_block[i] = sb.S_blocks_count
							err = fileBlock.Serialize(path, int64(sb.S_first_blo))
							if err != nil {
								return err
							}
							err = sb.UpdateBitmapBlock(path)
							if err != nil {
								return err
							}
							sb.S_blocks_count++
							sb.S_free_blocks_count--
							sb.S_first_blo += sb.S_block_size
						}
						fileInode.I_size = int32(len(finalContent))
					}

					// Serializar el inodo del archivo
					err = fileInode.Serialize(path, int64(sb.S_first_ino))
					if err != nil {
						return err
					}
					err = sb.UpdateBitmapInode(path)
					if err != nil {
						return err
					}
					sb.S_inodes_count++
					sb.S_free_inodes_count--
					sb.S_first_ino += sb.S_inode_size

					return nil
				}
			}
		}
	}
	return fmt.Errorf("no hay espacio en el directorio padre para crear %s", fileName)
}
