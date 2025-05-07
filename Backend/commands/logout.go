package commands

import (
	"backend/stores"
	"errors"
	"fmt"
)

func ParseLogout(tokens []string) (string, error) {
	// Validar que no se hayan enviado parámetros
	if len(tokens) > 0 {
		return "", errors.New("el comando logout no acepta parámetros")
	}

	// Verificar si hay una sesión activa
	if !stores.Auth.IsAuthenticated() {
		return "", errors.New("no hay una sesión activa para cerrar")
	}

	// Cerrar sesión
	stores.Auth.Logout()

	return fmt.Sprintf("Sesión cerrada exitosamente."), nil
}
