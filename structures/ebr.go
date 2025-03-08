package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type EBR struct {
	Part_status [1]byte  // Estado: 'N' (no usada), '0' (creada), '1' (montada)
	Part_fit    [1]byte  // Ajuste: 'B', 'F', 'W'
	Part_start  int32    // Byte de inicio
	Part_size   int32    // Tamaño en bytes
	Part_next   int32    // Byte de inicio del siguiente EBR, o -1 si no hay más
	Part_name   [16]byte // Nombre de la partición
}

// Serialize escribe el EBR en el archivo en la posición especificada
func (ebr *EBR) Serialize(file *os.File, offset int64) error {
	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("error al mover puntero: %v", err)
	}
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, ebr); err != nil {
		return fmt.Errorf("error al serializar EBR: %v", err)
	}
	_, err := file.Write(buffer.Bytes())
	return err
}

// Deserialize lee un EBR desde el archivo en la posición especificada
func (ebr *EBR) Deserialize(file *os.File, offset int64) error {
	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("error al mover puntero: %v", err)
	}
	return binary.Read(file, binary.LittleEndian, ebr)
}

// PrintEBR imprime los valores del EBR para depuración
func (ebr *EBR) PrintEBR() {
	fmt.Printf("Status: %c\n", ebr.Part_status[0])
	fmt.Printf("Fit: %c\n", ebr.Part_fit[0])
	fmt.Printf("Start: %d\n", ebr.Part_start)
	fmt.Printf("Size: %d\n", ebr.Part_size)
	fmt.Printf("Next: %d\n", ebr.Part_next)
	fmt.Printf("Name: %s\n", string(ebr.Part_name[:]))
}
