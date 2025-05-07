package services

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"

	admindiscos "Gestor/Comandos/AdminDiscos"
	filesFolders "Gestor/Comandos/AdminFiles"
	users "Gestor/Comandos/AdminUserGroup"
	rep "Gestor/Comandos/Rep"
	fileSystem "Gestor/Comandos/adminSistemaArchivos"

	"Gestor/models"
)

// AnalizarComando procesa un comando y devuelve el resultado
func AnalizarComando(entrada string) models.ResultadoComando {
	return analizarEntrada(entrada)
}

// GetLineasComando divide un texto en líneas individuales de comandos
func GetLineasComando(texto string) []string {
	var lineas []string
	scanner := bufio.NewScanner(strings.NewReader(texto))
	for scanner.Scan() {
		linea := strings.Split(scanner.Text(), "#")[0]
		if strings.TrimSpace(linea) != "" {
			lineas = append(lineas, linea)
		}
	}
	return lineas
}

// analizarEntrada divide y ejecuta un comando completo
func analizarEntrada(entrada string) models.ResultadoComando {
	parametros := parseParametros(entrada)
	resultado := models.ResultadoComando{Comando: entrada, Exito: true}

	var salida strings.Builder
	var errores strings.Builder

	if len(parametros) == 0 {
		errores.WriteString("ERROR: No se proporcionó ningún comando\n")
		return models.ResultadoComando{Comando: entrada, Exito: false, Errores: errores.String()}
	}

	cmd := strings.ToLower(parametros[0])

	// Función auxiliar para comandos estándar
	execute := func(etiqueta string, fn func([]string) string) {
		salida.WriteString(fmt.Sprintf("\n ========== %s ==========\n", strings.ToUpper(etiqueta)))
		if len(parametros) > 1 {
			res := fn(parametros)
			if strings.Contains(strings.ToLower(res), "error") {
				errores.WriteString(fmt.Sprintf("Error en comando: %s\n%s\n", entrada, res))
				resultado.Exito = false
			} else {
				salida.WriteString(fmt.Sprintf("Comando Ejecutado: %s\n%s\n", entrada, res))
			}
		} else {
			errores.WriteString(fmt.Sprintf("ERROR [%s]: faltan parámetros obligatorios\n", strings.ToUpper(etiqueta)))
			resultado.Exito = false
		}
		salida.WriteString(fmt.Sprintf(" ======= FIN %s ========\n\n", strings.ToUpper(etiqueta)))
	}

	// Comandos
	switch cmd {
	case "mkdisk":
		execute("mkdisk", admindiscos.Mkdisk)
	case "rmdisk":
		execute("rmdisk", admindiscos.Rmdisk)
	case "fdisk":
		execute("fdisk", admindiscos.Fdisk)
	case "mount":
		execute("mount", admindiscos.Mount)
	case "mounted":
		if len(parametros) == 1 {
			salida.WriteString(admindiscos.Mounted(parametros))
		} else {
			errores.WriteString("ERROR [MOUNTED]: no permite parámetros adicionales\n")
			resultado.Exito = false
		}
	case "mkfs":
		execute("mkfs", fileSystem.Mkfs)
	case "cat":
		execute("cat", fileSystem.Cat)
	case "login":
		execute("login", users.Login)
	case "logout":
		if len(parametros) == 1 {
			salida.WriteString(users.Logout(parametros))
		} else {
			errores.WriteString("ERROR [LOGOUT]: no permite parámetros adicionales\n")
			resultado.Exito = false
		}
	case "mkgrp":
		execute("mkgrp", users.Mkgrp)
	case "rmgrp":
		execute("rmgrp", users.Rmgrp)
	case "mkusr":
		execute("mkusr", users.Mkusr)
	case "rmusr":
		execute("rmusr", users.Rmusr)
	case "chgrp":
		execute("chgrp", users.Chgrp)
	case "mkdir":
		execute("mkdir", filesFolders.Mkdir)
	case "mkfile":
		execute("mkfile", filesFolders.Mkfile)
	case "rep":
		execute("rep", rep.Reportes)
	default:
		errores.WriteString(fmt.Sprintf("ERROR: comando no reconocido [%s]\n", cmd))
		resultado.Exito = false
	}

	resultado.Salida = salida.String()
	resultado.Errores = errores.String()

	return resultado
}

// parseParametros separa parámetros respetando comillas
func parseParametros(entrada string) []string {
	var parametros []string
	var buffer strings.Builder
	enComillas := false

	for i, char := range entrada {
		if char == '"' {
			enComillas = !enComillas
		}
		if char == '-' && !enComillas && i > 0 {
			parametros = append(parametros, buffer.String())
			buffer.Reset()
			continue
		}
		if !enComillas && unicode.IsSpace(char) {
			continue
		}
		buffer.WriteRune(char)
	}
	if buffer.Len() > 0 {
		parametros = append(parametros, buffer.String())
	}
	return parametros
}
