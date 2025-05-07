package AdminUserGroup

import (
	"Gestor/Estructuras"
	"Gestor/utils"
	"fmt"
)

/*
Comando para cerrar sesion
*/
func Logout(parametros []string) string {
	// 1) estructura para respuesta
	logger := utils.NewLogger("logput")
	// Encabezado
	logger.LogInfo("[ LOGOUT ]")
	// 2) verificar paramtros

	// 3) validar logica del comando LOGOUT
	usuario := Estructuras.UsuarioActual

	fmt.Println(" ----> Estado de la sesion: ", Estructuras.UsuarioActual.Status)
	// Para utilizar este comando es obligatorio que un usuario tenga una sesion abierta
	//validar que haya un usuario logeado
	if !usuario.Status {
		logger.LogError("ERROR [ LOGOUT ]: Actualmente no hay ninguna sesion abierta")
		return logger.GetErrors()
	} else {
		Estructuras.UsuarioActual.Status = false
		logger.LogInfo("Comando ejecutandose, cerrando sesion para  \" %s \"... \nSesion cerrada correctamete. ", usuario.Nombre)
	}

	fmt.Println(" ----> Estado de la sesion: ", Estructuras.UsuarioActual.Status)

	// 4) validar return salida
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()

}
