package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// LOGIN estructura que representa el comando login con sus parámetros
type LOGIN struct {
	user string // Usuario
	pass string // Contraseña
	id   string // ID del disco
}

/*
	login -user=root -pass=123 -id=062A3E2D
*/

func ParseLogin(tokens []string) (string, error) {
	cmd := &LOGIN{} // Crea una nueva instancia de LOGIN

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkfs
	re := regexp.MustCompile(`-user=[^\s]+|-pass=[^\s]+|-id=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)





	// Verificar que todos los tokens fueron reconocidos por la expresión regular
	if len(matches) != len(tokens) {
		// Identificar el parámetro inválido
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-user":
			if value == "" {
				return "", errors.New("el usuario no puede estar vacío")
			}
			cmd.user = value
		case "-pass":
			if value == "" {
				return "", errors.New("la contraseña no puede estar vacía")
			}
			cmd.pass = value
		case "-id":
			// Verifica que el id no esté vacío
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -id haya sido proporcionado
	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	// Si no se proporcionó el tipo, se establece por defecto a "full"
	if cmd.user == "" {
		return "", errors.New("faltan parámetros requeridos: -user")
	}

	// Si no se proporcionó el tipo, se establece por defecto a "full"
	if cmd.pass == "" {
		return "", errors.New("faltan parámetros requeridos: -pass")
	}

	// Aquí se puede agregar la lógica para ejecutar el comando mkfs con los parámetros proporcionados
	err := commandLogin(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"\n=============================LOGIN==============================\n"+
		"LOGIN: Usuario: %s, Contraseña: %s, ID: %s\n"+
		"==================================================================",
		cmd.user, cmd.pass, cmd.id), nil
	
}

func commandLogin(login *LOGIN) error {
	if stores.Auth.IsAuthenticated() {
		return fmt.Errorf("ya hay un usuario logueado, debe hacer logout primero")
	}

	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(login.id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	usersBlock, err := partitionSuperblock.GetUsersBlock(partitionPath)
	if err != nil {
		return fmt.Errorf("error al obtener el bloque de usuarios: %w", err)
	}

	content := strings.Trim(string(usersBlock.B_content[:]), "\x00")
	lines := strings.Split(content, "\n")

	var foundUser bool
	var uid, gid int
	var userPassword string

	for _, line := range lines {
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		if len(fields) == 5 && fields[1] == "U" {
			// Formato esperado: UID, U, grupo, usuario, contraseña
			if strings.EqualFold(fields[3], login.user) {
				foundUser = true
				userPassword = fields[4]

				// Convertir UID (fields[0]) y buscar GID por grupo
				uidParsed, err := strconv.Atoi(fields[0])
				if err != nil {
					return fmt.Errorf("error al convertir UID del usuario")
				}
				uid = uidParsed

				// Buscar GID (busca grupo por nombre)
				for _, gline := range lines {
					gfields := strings.Split(gline, ",")
					if len(gfields) == 3 && gfields[1] == "G" && strings.EqualFold(gfields[2], fields[2]) {
						gidParsed, err := strconv.Atoi(gfields[0])
						if err != nil {
							return fmt.Errorf("error al convertir GID del grupo")
						}
						gid = gidParsed
						break
					}
				}
				break
			}
		}
	}

	if !foundUser {
		return fmt.Errorf("el usuario %s no existe", login.user)
	}

	if !strings.EqualFold(userPassword, login.pass) {
		return fmt.Errorf("la contraseña no coincide")
	}

	// Guardar sesión
	stores.Auth.Login(login.user, login.pass, login.id, uid, gid)
	return nil
}
