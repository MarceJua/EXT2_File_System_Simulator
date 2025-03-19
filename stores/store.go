package stores

import (
	"errors"
	"fmt"
	"os"

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/structures"
)

// Carnet de estudiante
const Carnet string = "67" // 202010367

// Declaración de variables globales
var MountedPartitions = make(map[string]string)

func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error) {
	path, exists := MountedPartitions[id]
	if !exists {
		return nil, nil, "", errors.New("partición no montada")
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, nil, "", fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	var mbr structures.MBR
	if err := mbr.Deserialize(path); err != nil {
		return nil, nil, "", fmt.Errorf("error deserializando MBR: %v", err)
	}

	var startOffset int64
	for _, p := range mbr.Mbr_partitions {
		if string(p.Part_id[:]) == id {
			startOffset = int64(p.Part_start)
			break
		}
	}

	if startOffset == 0 {
		var extPartition *structures.Partition
		for _, p := range mbr.Mbr_partitions {
			if p.Part_type[0] == 'E' && p.Part_status[0] != 'N' {
				extPartition = &p
				break
			}
		}
		if extPartition == nil {
			return nil, nil, "", errors.New("no hay partición extendida")
		}

		var currentEBR structures.EBR
		currentOffset := int64(extPartition.Part_start)
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, nil, "", fmt.Errorf("error obteniendo tamaño del archivo: %v", err)
		}
		fileSize := fileInfo.Size()

		for currentOffset < fileSize {
			if err := currentEBR.Deserialize(file, currentOffset); err != nil {
				return nil, nil, "", fmt.Errorf("error leyendo EBR en offset %d: %v", currentOffset, err)
			}
			currentEBR.Print() // Depuración
			if string(currentEBR.Part_id[:]) == id {
				startOffset = int64(currentEBR.Part_start)
				break
			}
			if currentEBR.Part_next == -1 {
				return nil, nil, "", errors.New("partición lógica no encontrada")
			}
			currentOffset = int64(currentEBR.Part_next)
		}
		if startOffset == 0 {
			return nil, nil, "", errors.New("partición lógica no encontrada en EBRs")
		}
	}

	var sb structures.SuperBlock
	if err := sb.Deserialize(path, startOffset); err != nil {
		return nil, nil, "", fmt.Errorf("error deserializando superbloque en offset %d: %v", startOffset, err)
	}

	return &mbr, &sb, path, nil
}

// GetMountedPartitionSuperblock obtiene el SuperBlock de la partición montada con el id especificado
func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.Partition, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	err := mbr.Deserialize(path)
	if err != nil {
		return nil, nil, "", err
	}

	// Buscar la partición con el id especificado
	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	// Crear una instancia de SuperBlock
	var sb structures.SuperBlock

	// Deserializar la estructura SuperBlock desde un archivo binario
	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, partition, path, nil
}
