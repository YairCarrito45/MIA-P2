# Estructura del MBR y sus tamaños en disco

## **MBR (Master Boot Record)**

La estructura del MBR se define como sigue:

```go
 type MBR struct {
     Mbr_tamanio int32             // Tamaño del disco en bytes
     Mbr_creation_date [19]byte    // Fecha y hora de creación del MBR
     Mbr_disk_signature int32      // Firma del disco (ID)
     Mbr_disk_fit [1]byte          // Tipo de ajuste del disco
     Mbr_partitions [4]Partition   // Particiones del MBR (4 particiones)
 }
```

### **Tamaño de cada campo en bytes**

| Campo                 | Tipo       | Tamaño (bytes) |
|-----------------------|-----------|---------------|
| `Mbr_tamanio`        | `int32`    | 4             |
| `Mbr_creation_date`  | `[19]byte` | 19            |
| `Mbr_disk_signature` | `int32`    | 4             |
| `Mbr_disk_fit`       | `[1]byte`  | 1             |
| `Mbr_partitions`     | `[4]Partition` | 4 × 35 = 140 |
| **Total**            |            | **168 bytes** |

---

## **Estructura de una Partición**

La estructura de cada partición se define así:

```go
 type Partition struct {
     Part_status [1]byte     // Estado de la partición
     Part_type [1]byte       // Tipo de partición (P/E)
     Part_fit [1]byte        // Ajuste de la partición
     Part_start int32        // Byte de inicio de la partición
     Part_size int32         // Tamaño de la partición
     Part_name [16]byte      // Nombre de la partición
     Part_correlative int32  // Correlativo de la partición
     Part_id [4]byte         // ID de la partición
 }
```

### **Tamaño de cada campo en bytes**

| Campo                | Tipo       | Tamaño (bytes) |
|----------------------|-----------|---------------|
| `Part_status`       | `[1]byte`  | 1             |
| `Part_type`         | `[1]byte`  | 1             |
| `Part_fit`          | `[1]byte`  | 1             |
| `Part_start`       | `int32`     | 4             |
| `Part_size`        | `int32`     | 4             |
| `Part_name`        | `[16]byte`  | 16            |
| `Part_correlative` | `int32`     | 4             |
| `Part_id`          | `[4]byte`   | 4             |
| **Total**          |             | **35 bytes**  |

---

## **Resumen de tamaños**

- **Tamaño de una partición:** 35 bytes
- **Cantidad de particiones en el MBR:** 4
- **Tamaño total de `Mbr_partitions`:** 4 × 35 = 140 bytes
- **Tamaño total de la estructura `MBR`:** **168 bytes**

