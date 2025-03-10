package reports

import (
	"fmt"
	"strings"
	"time"

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/structures"
)

// ReportInode genera un reporte de un inodo y lo guarda en la ruta especificada
func ReportInode(sb *structures.SuperBlock, diskPath string) (string, error) {
	var sbBuilder strings.Builder
	sbBuilder.WriteString("digraph G {\n")
	sbBuilder.WriteString("  node [shape=plaintext]\n")

	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return "", fmt.Errorf("error deserializando inodo %d: %v", i, err)
		}

		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

		sbBuilder.WriteString(fmt.Sprintf("  inode%d [label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n", i))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD COLSPAN=\"2\">REPORTE INODO %d</TD></TR>\n", i))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_uid</TD><TD>%d</TD></TR>\n", inode.I_uid))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_gid</TD><TD>%d</TD></TR>\n", inode.I_gid))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_size</TD><TD>%d</TD></TR>\n", inode.I_size))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_atime</TD><TD>%s</TD></TR>\n", atime))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_ctime</TD><TD>%s</TD></TR>\n", ctime))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_mtime</TD><TD>%s</TD></TR>\n", mtime))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_type</TD><TD>%c</TD></TR>\n", inode.I_type[0]))
		sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>i_perm</TD><TD>%s</TD></TR>\n", string(inode.I_perm[:])))
		sbBuilder.WriteString("    <TR><TD COLSPAN=\"2\">BLOQUES DIRECTOS</TD></TR>\n")
		for j := 0; j < 12; j++ {
			sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>%d</TD><TD>%d</TD></TR>\n", j+1, inode.I_block[j]))
		}
		sbBuilder.WriteString("    <TR><TD COLSPAN=\"2\">BLOQUES INDIRECTOS</TD></TR>\n")
		for j := 12; j < 15; j++ {
			sbBuilder.WriteString(fmt.Sprintf("    <TR><TD>%d</TD><TD>%d</TD></TR>\n", j+1, inode.I_block[j]))
		}
		sbBuilder.WriteString("  </TABLE>>];\n")

		if i < sb.S_inodes_count-1 {
			sbBuilder.WriteString(fmt.Sprintf("  inode%d -> inode%d;\n", i, i+1))
		}
	}

	sbBuilder.WriteString("}\n")
	return sbBuilder.String(), nil
}
