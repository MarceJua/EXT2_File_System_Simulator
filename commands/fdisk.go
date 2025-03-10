package commands

import (
	"encoding/binary"
	"errors" // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"    // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"os"
	"regexp"  // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv" // Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas

	structures "github.com/MarceJua/MIA_1S2025_P1_202010367/structures"
	utils "github.com/MarceJua/MIA_1S2025_P1_202010367/utils"
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
func ParseFdisk(tokens []string) (*FDISK, error) {
	cmd := &FDISK{} // Crea una nueva instancia de FDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando fdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfF]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			if value != "K" && value != "M" {
				return nil, errors.New("la unidad debe ser K o M")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			// Verifica que el tipo sea "P", "E" o "L"
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return nil, errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			// Verifica que el nombre no esté vacío
			if value == "" {
				return nil, errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.size == 0 {
		return nil, errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -name")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if cmd.unit == "" {
		cmd.unit = "M"
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "FF"
	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	// Si no se proporcionó el tipo, se establece por defecto a "P"
	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Crear la partición con los parámetros proporcionados
	err := commandFdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil // Devuelve el comando FDISK creado
}

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

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	// Crear una instancia de MBR
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

	// Mensaje de éxito
	fmt.Printf("Partición primaria creada: %s\n", fdisk.name)
	return nil
}

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
		fmt.Println("Partición lógica creada:", fdisk.name)
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

	fmt.Println("Partición lógica creada:", fdisk.name)
	return nil
}

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

	// Crear la partición extendida
	partition.CreatePartition(start, sizeBytes, "E", fdisk.fit, fdisk.name)
	mbr.Mbr_partitions[idx] = *partition
	if err := mbr.Serialize(fdisk.path); err != nil {
		return err
	}
	fmt.Println("Partición extendida creada:", fdisk.name)
	return nil
}
