package reports

import (
	"fmt"
	"os"
	"strings"

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/structures"
)

func ReportBlock(sb *structures.SuperBlock, diskPath string) (string, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return "", fmt.Errorf("error al abrir disco: %v", err)
	}
	defer file.Close()

	var sbBuilder strings.Builder
	sbBuilder.WriteString("digraph G {\n")
	sbBuilder.WriteString("  node [shape=plaintext]\n")

	inodeSize := int(sb.S_inode_size) // 88 bytes
	blockSize := int(sb.S_block_size) // 64 bytes

	// Recorrer todos los inodos
	blockCounter := 0 // Para numerar los nodos en el grafo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(i*int32(inodeSize))))
		if err != nil {
			return "", fmt.Errorf("error deserializando inodo %d: %v", i, err)
		}

		// Procesar cada bloque directo del inodo (0 a 11)
		for j := 0; j < 12; j++ { // Solo bloques directos
			blockNum := inode.I_block[j]
			if blockNum == -1 {
				continue // Bloque no asignado
			}
			blockOffset := int64(sb.S_block_start + (blockNum * int32(blockSize)))

			if inode.I_type[0] == '0' { // Carpeta
				folderBlock := &structures.FolderBlock{}
				err = folderBlock.Deserialize(diskPath, blockOffset)
				if err != nil {
					return "", fmt.Errorf("error deserializando bloque carpeta %d: %v", blockNum, err)
				}
				hasContent := false
				for _, content := range folderBlock.B_content {
					if content.B_inodo != -1 && strings.TrimRight(string(content.B_name[:]), "\x00") != "" {
						hasContent = true
						break
					}
				}
				if hasContent {
					sbBuilder.WriteString(fmt.Sprintf("  block%d [label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n", blockCounter))
					sbBuilder.WriteString(fmt.Sprintf("    <TR><TD COLSPAN=\"2\">Bloque Carpeta %d</TD></TR>\n", blockNum))
					sbBuilder.WriteString("    <TR><TD>b_name</TD><TD>b_inodo</TD></TR>\n")
					for _, content := range folderBlock.B_content {
						name := strings.TrimRight(string(content.B_name[:]), "\x00")
						if name != "" && content.B_inodo != -1 {
							// Escapar caracteres especiales
							name = strings.ReplaceAll(name, "<", "&lt;")
							name = strings.ReplaceAll(name, ">", "&gt;")
							name = strings.ReplaceAll(name, "&", "&amp;")
							sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>%s</TD><TD>%d</TD></TR>\n", name, content.B_inodo))
						}
					}
					sbBuilder.WriteString("  </TABLE>>];\n")
					blockCounter++
				}
			} else if inode.I_type[0] == '1' { // Archivo
				fileBlock := &structures.FileBlock{}
				err = fileBlock.Deserialize(diskPath, blockOffset)
				if err != nil {
					return "", fmt.Errorf("error deserializando bloque archivo %d: %v", blockNum, err)
				}
				content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
				if content != "" { // Solo mostrar bloques con contenido
					// Escapar caracteres especiales y reemplazar saltos de línea
					content = strings.ReplaceAll(content, "<", "&lt;")
					content = strings.ReplaceAll(content, ">", "&gt;")
					content = strings.ReplaceAll(content, "&", "&amp;")
					content = strings.ReplaceAll(content, "\n", "<BR/>") // Usar <BR/> para saltos de línea en HTML
					sbBuilder.WriteString(fmt.Sprintf("  block%d [label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n", blockCounter))
					sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>Bloque Archivo %d</TD></TR>\n", blockNum))
					sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>%s</TD></TR>\n", content))
					sbBuilder.WriteString("  </TABLE>>];\n")
					blockCounter++
				}
			}
		}
	}

	if blockCounter == 0 {
		sbBuilder.WriteString("  node0 [label=\"No hay bloques usados\"];\n")
	}

	sbBuilder.WriteString("}\n")
	return sbBuilder.String(), nil
}
