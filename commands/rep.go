package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	reports "github.com/MarceJua/MIA_1S2025_P1_202010367/reports"
	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/stores"
)

// REP estructura que representa el comando rep con sus parámetros
type REP struct {
	id           string // ID del disco
	path         string // Ruta del archivo del disco
	name         string // Nombre del reporte
	path_file_ls string // Ruta del archivo ls (opcional)
}

// ParserRep parsea el comando rep y devuelve una instancia de REP
func ParseRep(tokens []string) (*REP, error) {
	cmd := &REP{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-path="[^"]+"|-path=[^\s]+|-name=[^\s]+|-path_file_ls="[^"]+"|-path_file_ls=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), strings.Trim(kv[1], "\"")
		switch key {
		case "-id":
			if value == "" {
				return nil, errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-path":
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			validNames := []string{"mbr", "ebr", "disk", "inode", "block", "bm_inode", "bm_block", "tree", "sb", "file", "ls"}
			if !contains(validNames, value) {
				return nil, errors.New("nombre inválido, debe ser: mbr, ebr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls")
			}
			cmd.name = value
		case "-path_file_ls":
			if value == "" {
				return nil, errors.New("el path_file_ls no puede estar vacío")
			}
			cmd.path_file_ls = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" || cmd.path == "" || cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -id, -path, -name")
	}
	if cmd.name == "ls" && cmd.path_file_ls == "" {
		return nil, errors.New("falta parámetro -path_file_ls para reporte ls")
	}

	err := commandRep(cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func commandRep(rep *REP) error {
	mountedMbr, mountedSb, mountedDiskPath, err := stores.GetMountedPartitionRep(rep.id)
	if err != nil {
		return err
	}

	// Generar el reporte según el tipo
	var dotContent string
	switch rep.name {
	case "mbr":
		dotContent, err = reports.ReportMBR(mountedMbr)
	case "ebr":
		dotContent, err = reports.ReportEBR(mountedMbr, mountedDiskPath)
	case "disk":
		dotContent, err = reports.ReportDisk(mountedMbr, mountedDiskPath)
	case "inode":
		dotContent, err = reports.ReportInode(mountedSb, mountedDiskPath)
	case "block":
		dotContent, err = reports.ReportBlock(mountedSb, mountedDiskPath)
	case "bm_inode":
		dotContent, err = reports.ReportBMInode(mountedSb, mountedDiskPath)
	case "bm_block":
		dotContent, err = reports.ReportBMBlock(mountedSb, mountedDiskPath)
	case "tree":
		dotContent, err = reports.ReportTree(mountedSb, mountedDiskPath)
	case "sb":
		dotContent, err = reports.ReportSB(mountedSb)
		//	case "file":
		//		dotContent, err = reports.ReportFile(mountedSb, mountedDiskPath)
		//	case "ls":
		//		dotContent, err = reports.ReportLS(mountedSb, mountedDiskPath, rep.path_file_ls)
	default:
		return fmt.Errorf("reporte no implementado: %s", rep.name)
	}
	if err != nil {
		return fmt.Errorf("error generando reporte %s: %v", rep.name, err)
	}

	// Escribir archivo DOT
	dotFile := rep.path + ".dot"
	err = writeDotFile(dotFile, dotContent)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo DOT: %v", err)
	}

	// Convertir a imagen con Graphviz
	outputFile := strings.TrimSuffix(rep.path, filepath.Ext(rep.path)) + ".png"
	err = generateImage(dotFile, outputFile)
	if err != nil {
		return fmt.Errorf("error generando imagen: %v", err)
	}

	fmt.Printf("Reporte %s generado en %s\n", rep.name, outputFile)
	return nil
}

func writeDotFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func generateImage(dotFile, outputFile string) error {
	cmd := exec.Command("dot", "-Tpng", dotFile, "-o", outputFile)
	return cmd.Run()
}
