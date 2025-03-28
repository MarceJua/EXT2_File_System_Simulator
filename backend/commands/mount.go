package commands

import (
	"errors" // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"    // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"os"     // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv"

	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/stores"
	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/structures" // Paquete que contiene las estructuras de datos necesarias para el manejo de discos y particiones

	// Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas
)

// MOUNT estructura que representa el comando mount con sus parámetros
type MOUNT struct {
	path string // Ruta del archivo del disco
	name string // Nombre de la partición
}

/*
	mount -path=/home/Disco1.mia -name=Part1 #id=341a
	mount -path=/home/Disco2.mia -name=Part1 #id=342a
	mount -path=/home/Disco3.mia -name=Part2 #id=343a
*/

// CommandMount parsea el comando mount y devuelve una instancia de MOUNT
func ParseMount(tokens []string) (string, error) {
	cmd := &MOUNT{}

	// Procesar cada token
	for _, token := range tokens {
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("formato inválido: %s", token)
		}
		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Validar parámetros requeridos
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	// Ejecutar el comando
	id, err := commandMount(cmd)
	if err != nil {
		return "", fmt.Errorf("error al montar la partición: %v", err)
	}

	return fmt.Sprintf("MOUNT: Partición %s montada correctamente con ID: %s", cmd.name, id), nil
}

// commandMount implementa la lógica para montar la partición y devuelve el ID asignado
func commandMount(mount *MOUNT) (string, error) {
	var mbr structures.MBR
	if err := mbr.Deserialize(mount.path); err != nil {
		return "", fmt.Errorf("error al deserializar MBR: %v", err)
	}

	// Contar particiones montadas para este disco y determinar el correlativo
	correlative := 1
	for id, path := range stores.MountedPartitions {
		if path == mount.path {
			num := id[2:3] // Extraer el correlativo (e.g., "1" de "671A")
			if n, err := strconv.Atoi(num); err == nil && n >= correlative {
				correlative = n + 1
			}
		}
	}

	// Generar el ID
	id := fmt.Sprintf("%s%dA", stores.Carnet, correlative)
	if _, exists := stores.MountedPartitions[id]; exists {
		return "", errors.New("el ID ya está en uso")
	}

	// Buscar en primarias
	partition, idx := mbr.GetPartitionByName(mount.name)
	if partition != nil {
		if partition.Part_status[0] == '1' {
			return "", errors.New("la partición ya está montada")
		}
		if partition.Part_type[0] == 'E' {
			return "", errors.New("no se pueden montar particiones extendidas")
		}

		partition.MountPartition(correlative, id)
		mbr.Mbr_partitions[idx] = *partition
		stores.MountedPartitions[id] = mount.path
		if err := mbr.Serialize(mount.path); err != nil {
			return "", fmt.Errorf("error al serializar MBR: %v", err)
		}
		return id, nil
	}

	// Buscar en lógicas
	file, err := os.OpenFile(mount.path, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("error al abrir disco: %v", err)
	}
	defer file.Close()

	var extPartition *structures.Partition
	for _, p := range mbr.Mbr_partitions {
		if p.Part_type[0] == 'E' && p.Part_status[0] != 'N' {
			extPartition = &p
			break
		}
	}
	if extPartition == nil {
		return "", errors.New("partición no encontrada (no hay extendida para lógicas)")
	}

	startExt := int64(extPartition.Part_start)
	var currentEBR structures.EBR
	err = currentEBR.Deserialize(file, startExt)
	if err != nil || currentEBR.Part_status[0] == 0 || currentEBR.Part_status[0] == 'N' {
		return "", errors.New("partición lógica no encontrada")
	}

	currentOffset := startExt
	for {
		ebName := strings.Trim(string(currentEBR.Part_name[:]), "\x00")
		if ebName == mount.name {
			if currentEBR.Part_status[0] == '1' {
				return "", errors.New("la partición lógica ya está montada")
			}

			currentEBR.Part_status = [1]byte{'1'}
			copy(currentEBR.Part_id[:], id)
			if err := currentEBR.Serialize(file, currentOffset); err != nil {
				return "", fmt.Errorf("error al serializar EBR: %v", err)
			}
			stores.MountedPartitions[id] = mount.path
			return id, nil
		}

		if currentEBR.Part_next == -1 {
			break
		}
		currentOffset = int64(currentEBR.Part_next)
		if err := currentEBR.Deserialize(file, currentOffset); err != nil {
			return "", fmt.Errorf("error al leer EBR: %v", err)
		}
	}

	return "", errors.New("partición lógica no encontrada")
}
