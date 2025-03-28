package commands

import (
	"encoding/binary"
	"errors"  // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"     // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"os"      // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv" // Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/structures"
	utils "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/utils"
)

// FDISK estructura que representa el comando fdisk con sus parámetros
type FDISK struct {
	size int    // Tamaño de la partición
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
	typ  string // Tipo de partición (P, E, L)
	name string // Nombre de la partición
}

/*
	fdisk -size=1 -type=L -unit=M -fit=BF -name="Particion3" -path="/home/keviin/University/PRACTICAS/MIA_LAB_S2_2024/CLASEEXTRA/disks/Disco1.mia"
	fdisk -size=300 -path=/home/Disco1.mia -name=Particion1
	fdisk -type=E -path=/home/Disco2.mia -Unit=K -name=Particion2 -size=300
*/

// CommandFdisk parsea el comando fdisk y devuelve una instancia de FDISK
func ParseFdisk(tokens []string) (string, error) {
	cmd := &FDISK{}

	// Procesar cada token
	for _, token := range tokens {
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("formato inválido: %s", token)
		}
		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" {
				return "", errors.New("la unidad debe ser K o M")
			}
			cmd.unit = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
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
	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	// Establecer valores por defecto
	if cmd.unit == "" {
		cmd.unit = "M"
	}
	if cmd.fit == "" {
		cmd.fit = "WF"
	}
	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Ejecutar el comando
	err := commandFdisk(cmd)
	if err != nil {
		return "", fmt.Errorf("error al crear la partición: %v", err)
	}

	return fmt.Sprintf("FDISK: Partición %s creada correctamente en %s", cmd.name, cmd.path), nil
}

// commandFdisk implementa la lógica para crear la partición
func commandFdisk(fdisk *FDISK) error {
	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		return fmt.Errorf("error al convertir tamaño: %v", err)
	}

	var mbr structures.MBR
	if err := mbr.Deserialize(fdisk.path); err != nil {
		return fmt.Errorf("error al deserializar MBR: %v", err)
	}

	// Validar nombre duplicado en primarias/extendidas
	if _, idx := mbr.GetPartitionByName(fdisk.name); idx != -1 {
		return fmt.Errorf("el nombre '%s' ya existe en particiones primarias/extendidas", fdisk.name)
	}

	switch fdisk.typ {
	case "P":
		return createPrimaryPartition(fdisk, sizeBytes)
	case "E":
		return createExtendedPartition(fdisk, sizeBytes)
	case "L":
		return createLogicalPartition(fdisk, sizeBytes)
	default:
		return errors.New("tipo de partición inválido")
	}
}

// createPrimaryPartition crea una partición primaria
func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR
	if err := mbr.Deserialize(fdisk.path); err != nil {
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// Contar particiones primarias/extendidas
	count := 0
	for _, p := range mbr.Mbr_partitions {
		if p.Part_status[0] != 'N' {
			count++
		}
	}
	if count >= 4 {
		return errors.New("máximo de 4 particiones primarias/extendidas alcanzado")
	}

	partition, start, idx := mbr.GetFirstAvailablePartition()
	if partition == nil {
		return errors.New("no hay particiones disponibles")
	}

	if sizeBytes > int(mbr.Mbr_size)-start {
		return errors.New("no hay espacio suficiente en el disco")
	}

	partition.CreatePartition(start, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)
	mbr.Mbr_partitions[idx] = *partition
	if err := mbr.Serialize(fdisk.path); err != nil {
		return fmt.Errorf("error serializando el MBR: %v", err)
	}

	return nil
}

// createLogicalPartition crea una partición lógica dentro de una extendida
func createLogicalPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR
	if err := mbr.Deserialize(fdisk.path); err != nil {
		return fmt.Errorf("error al deserializar MBR: %v", err)
	}

	var extPartition *structures.Partition
	for _, p := range mbr.Mbr_partitions {
		if p.Part_type[0] == 'E' && p.Part_status[0] != 'N' {
			extPartition = &p
			break
		}
	}
	if extPartition == nil {
		return errors.New("no hay partición extendida para crear lógicas")
	}

	file, err := os.OpenFile(fdisk.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer file.Close()

	startExt := int64(extPartition.Part_start)
	availableSpace := int(extPartition.Part_size)

	var currentEBR structures.EBR
	err = currentEBR.Deserialize(file, startExt)
	if err != nil || currentEBR.Part_status[0] == 0 || currentEBR.Part_status[0] == 'N' {
		if sizeBytes > availableSpace {
			return errors.New("no hay espacio suficiente en la partición extendida")
		}
		currentEBR = structures.EBR{
			Part_status: [1]byte{'0'},
			Part_fit:    [1]byte{fdisk.fit[0]},
			Part_start:  extPartition.Part_start,
			Part_size:   int32(sizeBytes),
			Part_next:   -1,
		}
		copy(currentEBR.Part_name[:], fdisk.name)
		if err := currentEBR.Serialize(file, startExt); err != nil {
			return fmt.Errorf("error al crear primer EBR: %v", err)
		}
		return nil
	}

	currentOffset := startExt
	for {
		if string(currentEBR.Part_name[:]) == fdisk.name {
			return fmt.Errorf("el nombre '%s' ya existe en particiones lógicas", fdisk.name)
		}
		if currentEBR.Part_next == -1 {
			break
		}
		currentOffset = int64(currentEBR.Part_next)
		if err := currentEBR.Deserialize(file, currentOffset); err != nil {
			return fmt.Errorf("error al leer EBR: %v", err)
		}
		availableSpace -= int(currentEBR.Part_size)
	}

	nextStart := currentOffset + int64(binary.Size(currentEBR))
	availableSpace -= int(nextStart - startExt)
	if sizeBytes > availableSpace {
		return errors.New("no hay espacio suficiente en la partición extendida")
	}

	newEBR := structures.EBR{
		Part_status: [1]byte{'0'},
		Part_fit:    [1]byte{fdisk.fit[0]},
		Part_start:  int32(nextStart),
		Part_size:   int32(sizeBytes),
		Part_next:   -1,
	}
	copy(newEBR.Part_name[:], fdisk.name)

	currentEBR.Part_next = int32(nextStart)
	if err := currentEBR.Serialize(file, currentOffset); err != nil {
		return fmt.Errorf("error al actualizar EBR anterior: %v", err)
	}
	if err := newEBR.Serialize(file, int64(newEBR.Part_start)); err != nil {
		return fmt.Errorf("error al crear nuevo EBR: %v", err)
	}

	return nil
}

// createExtendedPartition crea una partición extendida
func createExtendedPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR
	if err := mbr.Deserialize(fdisk.path); err != nil {
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// Validar que no exista otra extendida
	for _, p := range mbr.Mbr_partitions {
		if p.Part_type[0] == 'E' && p.Part_status[0] != 'N' {
			return errors.New("ya existe una partición extendida en el disco")
		}
	}

	// Contar particiones primarias/extendidas
	count := 0
	for _, p := range mbr.Mbr_partitions {
		if p.Part_status[0] != 'N' {
			count++
		}
	}
	if count >= 4 {
		return errors.New("máximo de 4 particiones primarias/extendidas alcanzado")
	}

	partition, start, idx := mbr.GetFirstAvailablePartition()
	if partition == nil {
		return errors.New("no hay particiones disponibles")
	}

	if sizeBytes > int(mbr.Mbr_size)-start {
		return errors.New("no hay espacio suficiente en el disco")
	}

	partition.CreatePartition(start, sizeBytes, "E", fdisk.fit, fdisk.name)
	mbr.Mbr_partitions[idx] = *partition
	if err := mbr.Serialize(fdisk.path); err != nil {
		return fmt.Errorf("error serializando el MBR: %v", err)
	}

	return nil
}
