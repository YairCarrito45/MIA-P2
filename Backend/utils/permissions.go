package utils




import (
	"backend/stores"
	"backend/structures"
)


// HasWritePermission verifica si el usuario actual tiene permiso de escritura en el inodo
func HasWritePermission(inode structures.Inode) bool {
	username, _, _ := stores.Auth.GetCurrentUser()

	// root siempre tiene todos los permisos
	if username == "root" {
		return true
	}

	perm := string(inode.I_perm[:]) // Ej: "764"

	var accessChar byte
	if int(inode.I_uid) == stores.Auth.UserID {
		accessChar = perm[0] // Usuario propietario
	} else if int(inode.I_gid) == stores.Auth.GroupID {
		accessChar = perm[1] // Grupo
	} else {
		accessChar = perm[2] // Otros
	}

	// Convertimos el caracter a permiso numérico y validamos bit de escritura
	// 0 = ---  (000), 1 = --x (001), 2 = -w- (010), 3 = -wx (011), ...
	return accessChar == '2' || accessChar == '3' || accessChar == '6' || accessChar == '7'
}