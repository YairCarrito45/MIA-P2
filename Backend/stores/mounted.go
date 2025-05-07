package stores

import (
	structures "backend/structures"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Carnet de estudiante
const Carnet string = "78"

// Información completa de una partición montada
type MountInfo struct {
	Path        string // Ruta del disco
	Name        string // Nombre de la partición
	Letter      string // Letra asignada al disco (A, B, C...)
	Correlative int    // Número de partición montada (1, 2, 3...)
}

// Declaración de variables globales
var (
	MountedPartitions map[string]MountInfo = make(map[string]MountInfo)
)

// GetMountedPartition obtiene la partición montada con el id especificado
func GetMountedPartition(id string) (*structures.Partition, string, error) {
	info, ok := MountedPartitions[id]
	if !ok {
		return nil, "", errors.New("la partición no está montada")
	}
	path := info.Path

	var mbr structures.MBR
	err := mbr.Deserialize(path)
	if err != nil {
		return nil, "", err
	}

	partition, _ := mbr.GetPartitionByName(info.Name)
	if partition == nil {
		return nil, "", errors.New("partición no encontrada")
	}

	return partition, path, nil
}

// GetMountedPartitionRep obtiene el MBR y SuperBlock de la partición montada
func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error) {
	info, exists := MountedPartitions[id]
	if !exists {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	path := info.Path

	mbr, err := structures.ReadMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, _ := mbr.GetPartitionByName(info.Name)
	if partition != nil {
		var sb structures.SuperBlock
		err := sb.Deserialize(path, int64(partition.Part_start))
		if err != nil {
			return nil, nil, "", err
		}
		return &mbr, &sb, path, nil
	}

	ebr, err := mbr.GetLogicalPartitionByName(info.Name, path)
	if err != nil {
		return nil, nil, "", errors.New("partición no encontrada")
	}

	var sb structures.SuperBlock
	err = sb.Deserialize(path, int64(ebr.PartStart))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil
}

// GetMountedPartitionSuperblock obtiene el SuperBlock y partición montada con el id
func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.Partition, string, error) {
	info, ok := MountedPartitions[id]
	if !ok {
		return nil, nil, "", errors.New("la partición no está montada")
	}
	path := info.Path

	var mbr structures.MBR
	err := mbr.Deserialize(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, _ := mbr.GetPartitionByName(info.Name)
	if partition != nil {
		var sb structures.SuperBlock
		err := sb.Deserialize(path, int64(partition.Part_start))
		if err != nil {
			return nil, nil, "", err
		}
		return &sb, partition, path, nil
	}

	ebr, err := mbr.GetLogicalPartitionByName(info.Name, path)
	if err != nil {
		return nil, nil, "", errors.New("partición no encontrada")
	}

	var sb structures.SuperBlock
	err = sb.Deserialize(path, int64(ebr.PartStart))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, nil, path, nil
}

// ShowMountedPartitions imprime los IDs de todas las particiones montadas
func ShowMountedPartitions() string {
	if len(MountedPartitions) == 0 {
		return "No hay particiones montadas."
	}

	grouped := make(map[string][]string)
	for id := range MountedPartitions {
		prefix := id[:len(id)-1]
		grouped[prefix] = append(grouped[prefix], id)
	}

	var prefixes []string
	for prefix := range grouped {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)

	var result strings.Builder
	result.WriteString("======================== MOUNTED =========================\n")
	result.WriteString("Particiones montadas:\n")

	for _, prefix := range prefixes {
		ids := grouped[prefix]
		sort.Strings(ids)
		result.WriteString("  Disco " + prefix + ": " + strings.Join(ids, ", ") + "\n")
	}
	result.WriteString("============================================================\n")
	return result.String()
}

//////////////////////////////////////////////////////////////
// NUEVO: Soporte para /disks
//////////////////////////////////////////////////////////////

type MountedDisk struct {
	Name              string   `json:"name"`
	Path              string   `json:"path"` // NUEVO
	Size              string   `json:"size"`
	Fit               string   `json:"fit"`
	MountedPartitions []string `json:"mounted_partitions"`
}

// Devuelve una lista de discos montados con info básica
func GetMountedDisks() []MountedDisk {
	diskMap := map[string]*MountedDisk{}

	for id, info := range MountedPartitions {
		path := info.Path

		mbr, err := structures.ReadMBR(path)
		if err != nil {
			continue
		}

		if _, exists := diskMap[path]; !exists {
			name := extractFileName(path)
			size := fmt.Sprintf("%d MB", mbr.Mbr_size/1024/1024)
			fit := string(mbr.Mbr_disk_fit[:])

			diskMap[path] = &MountedDisk{
				Name:              name,
				Path:              path, // NUEVO
				Size:              size,
				Fit:               fit,
				MountedPartitions: []string{},
			}
		}

		diskMap[path].MountedPartitions = append(diskMap[path].MountedPartitions, id)
	}

	var result []MountedDisk
	for _, d := range diskMap {
		result = append(result, *d)
	}

	return result
}

func extractFileName(path string) string {
	splits := strings.Split(path, "/")
	return splits[len(splits)-1]
}
