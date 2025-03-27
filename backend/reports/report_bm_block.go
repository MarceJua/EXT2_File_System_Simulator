package reports

import (
	"fmt"
	"os"
	"strings"

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/structures"
)

func ReportBMBlock(sb *structures.SuperBlock, diskPath string) (string, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	_, err = file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil {
		return "", fmt.Errorf("error buscando bitmap de bloques: %v", err)
	}

	totalBlocks := sb.S_blocks_count + sb.S_free_blocks_count
	buffer := make([]byte, totalBlocks)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de bloques: %v", err)
	}

	var sbBuilder strings.Builder
	sbBuilder.WriteString("digraph G {\n")
	sbBuilder.WriteString("  node [shape=plaintext]\n")
	sbBuilder.WriteString("  tbl [label=<<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	sbBuilder.WriteString("    <TR><TD><B>Bitmap Bloques</B></TD></TR>\n")

	for i, bit := range buffer {
		if i%20 == 0 {
			sbBuilder.WriteString("    <TR>")
		}
		sbBuilder.WriteString(fmt.Sprintf("<TD>%c</TD>", bit))
		if (i+1)%20 == 0 || i == len(buffer)-1 {
			sbBuilder.WriteString("</TR>\n")
		}
	}

	sbBuilder.WriteString("  </TABLE>>];\n")
	sbBuilder.WriteString("}\n")
	return sbBuilder.String(), nil
}
