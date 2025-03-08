package commands

import (
	"errors" // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"    // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"os"
	"regexp" // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas

	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/stores"
	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/structures" // Paquete que contiene las estructuras de datos necesarias para el manejo de discos y particiones
	utils "github.com/MarceJua/MIA_1S2025_P1_202010367/utils"

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
func ParseMount(tokens []string) (*MOUNT, error) {
	cmd := &MOUNT{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key, value := strings.ToLower(kv[0]), strings.Trim(kv[1], "\"")
		switch key {
		case "-path":
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return nil, errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		}
	}

	if cmd.path == "" || cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -path o -name")
	}

	err := commandMount(cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func commandMount(mount *MOUNT) error {
	// Crear una instancia de MBR
	var mbr structures.MBR
	if err := mbr.Deserialize(mount.path); err != nil {
		return fmt.Errorf("error al deserializar MBR: %v", err)
	}

	// Buscar en primarias/extendidas primero
	partition, idx := mbr.GetPartitionByName(mount.name)
	if partition != nil {
		if partition.Part_status[0] == '1' {
			return errors.New("la partición ya está montada")
		}
		if partition.Part_type[0] == 'E' {
			return errors.New("no se pueden montar particiones extendidas")
		}

		id, correlative, err := generatePartitionID(mount)
		if err != nil {
			return fmt.Errorf("error generando ID: %v", err)
		}
		if _, exists := stores.MountedPartitions[id]; exists {
			return errors.New("el ID ya está en uso")
		}

		partition.MountPartition(correlative, id)
		mbr.Mbr_partitions[idx] = *partition
		stores.MountedPartitions[id] = mount.path
		fmt.Printf("Partición primaria montada con ID: %s\n", id)
		return mbr.Serialize(mount.path)
	}

	// Si no está en el MBR, buscar en las lógicas
	file, err := os.OpenFile(mount.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer file.Close()

	// Buscar partición extendida
	var extPartition *structures.Partition
	for _, p := range mbr.Mbr_partitions {
		if p.Part_type[0] == 'E' && p.Part_status[0] != 'N' {
			extPartition = &p
			break
		}
	}
	if extPartition == nil {
		return errors.New("partición no encontrada (no hay extendida para lógicas)")
	}

	// Recorrer los EBRs
	startExt := int64(extPartition.Part_start)
	var currentEBR structures.EBR
	err = currentEBR.Deserialize(file, startExt)
	if err != nil || currentEBR.Part_status[0] == 0 || currentEBR.Part_status[0] == 'N' {
		return errors.New("partición lógica no encontrada")
	}

	currentOffset := startExt
	for {
		ebName := strings.Trim(string(currentEBR.Part_name[:]), "\x00")
		if ebName == mount.name {
			if currentEBR.Part_status[0] == '1' {
				return errors.New("la partición lógica ya está montada")
			}

			id, _, err := generatePartitionID(mount) // Ignorar correlative con _
			if err != nil {
				return fmt.Errorf("error generando ID: %v", err)
			}
			if _, exists := stores.MountedPartitions[id]; exists {
				return errors.New("el ID ya está en uso")
			}

			// Actualizar EBR a montada
			currentEBR.Part_status = [1]byte{'1'}
			if err := currentEBR.Serialize(file, currentOffset); err != nil {
				return fmt.Errorf("error al serializar EBR: %v", err)
			}
			stores.MountedPartitions[id] = mount.path
			fmt.Printf("Partición lógica montada con ID: %s\n", id)
			return nil
		}

		if currentEBR.Part_next == -1 {
			break
		}
		currentOffset = int64(currentEBR.Part_next)
		if err := currentEBR.Deserialize(file, currentOffset); err != nil {
			return fmt.Errorf("error al leer EBR: %v", err)
		}
	}

	return errors.New("partición lógica no encontrada")
}

func generatePartitionID(mount *MOUNT) (string, int, error) {
	// Asignar una letra a la partición y obtener el índice
	letter, partitionCorrelative, err := utils.GetLetterAndPartitionCorrelative(mount.path)
	if err != nil {
		return "", 0, fmt.Errorf("error obteniendo letra: %v", err)
	}
	idPartition := fmt.Sprintf("%s%d%s", stores.Carnet, partitionCorrelative, letter)
	return idPartition, partitionCorrelative, nil
}
