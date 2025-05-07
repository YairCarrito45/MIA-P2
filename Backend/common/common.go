package common

import (
	"errors"
	"sort"
)

type PartPos struct {
	Start int64
	End   int64
	Type  byte
	Name  string
}

// Cualquier función general que no use directamente structures específicas:
func SortPartsByStart(parts []PartPos) {
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].Start < parts[j].Start
	})
}


// First devuelve el primer elemento de un slice
func First[T any](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, errors.New("el slice está vacío")
	}
	return slice[0], nil
}


// RemoveElement elimina un elemento de un slice en el índice dado
func RemoveElement[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice // Índice fuera de rango, devolver el slice original
	}
	return append(slice[:index], slice[index+1:]...)
}