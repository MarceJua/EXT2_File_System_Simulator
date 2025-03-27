package reports

import (
	"fmt"
	"os"
	"strings"

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/structures"
)

func ReportTree(sb *structures.SuperBlock, diskPath string) (string, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return "", fmt.Errorf("error al abrir disco: %v", err)
	}
	defer file.Close()

	var sbBuilder strings.Builder
	sbBuilder.WriteString("digraph Tree {\n")
	sbBuilder.WriteString("  node [shape=box]\n")

	inodeSize := int(sb.S_inode_size) // 88 bytes
	blockSize := int(sb.S_block_size) // 64 bytes

	// Mapa para evitar duplicados
	processedInodes := make(map[int32]bool)

	// Función recursiva para construir el árbol
	var buildTree func(inodoNum int32, parentName string) error
	buildTree = func(inodoNum int32, parentName string) error {
		if processedInodes[inodoNum] {
			return nil // Evitar ciclos o duplicados
		}
		processedInodes[inodoNum] = true

		// Leer el inodo
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(inodoNum*int32(inodeSize))))
		if err != nil {
			return fmt.Errorf("error deserializando inodo %d: %v", inodoNum, err)
		}

		// Procesar el primer bloque directo (I_block[0])
		blockNum := inode.I_block[0]
		if blockNum == -1 {
			return nil // Sin bloque asignado
		}
		blockOffset := int64(sb.S_block_start + (blockNum * int32(blockSize)))

		if inode.I_type[0] == '0' { // Carpeta
			folderBlock := &structures.FolderBlock{}
			err = folderBlock.Deserialize(diskPath, blockOffset)
			if err != nil {
				return fmt.Errorf("error deserializando bloque carpeta %d: %v", blockNum, err)
			}

			// Nombre del nodo actual (raíz es "/")
			currentName := "\"/\""
			if parentName != "" {
				currentName = parentName // Usar el nombre asignado por el padre
			}

			// Procesar cada entrada en el FolderBlock
			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				childInodo := content.B_inodo
				if name != "" && childInodo != -1 && name != "." && name != ".." {
					// Nombre del hijo
					childName := fmt.Sprintf("\"%s\"", name)
					// Conectar carpeta con su contenido
					sbBuilder.WriteString(fmt.Sprintf("  %s -> %s\n", currentName, childName))
					// Recursivamente procesar el inodo hijo
					err = buildTree(childInodo, childName)
					if err != nil {
						return err
					}
				}
			}
		}
		// Archivos no tienen hijos, solo se conectan desde el padre
		return nil
	}

	// Comenzar desde el inodo raíz (0)
	err = buildTree(0, "")
	if err != nil {
		return "", err
	}

	sbBuilder.WriteString("}\n")
	return sbBuilder.String(), nil
}
