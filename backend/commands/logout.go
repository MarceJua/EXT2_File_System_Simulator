package commands

import (
	"errors"

	stores "github.com/MarceJua/MIA_1S2025_P1_202010367/backend/stores"
)

// LOGOUT estructura que representa el comando logout (sin parámetros)
type LOGOUT struct{}

// ParseLogout parsea los tokens del comando logout
func ParseLogout(tokens []string) (string, error) {
	// Verificar que no haya parámetros adicionales
	if len(tokens) > 1 {
		return "", errors.New("el comando logout no acepta parámetros")
	}

	// Ejecutar el comando
	err := commandLogout()
	if err != nil {
		return "", err
	}

	return "LOGOUT: Sesión cerrada correctamente", nil
}

// commandLogout implementa la lógica del comando logout
func commandLogout() error {
	// Verificar si hay una sesión activa
	if stores.CurrentSession.ID == "" {
		return errors.New("no hay ninguna sesión activa para cerrar")
	}

	// Limpiar la sesión
	stores.CurrentSession = stores.Session{
		ID:       "",
		Username: "",
		UID:      "",
		GID:      "",
	}

	return nil
}
