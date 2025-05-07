

## Visión general del disco con particiones

Primero, veamos cómo se organiza un disco completo:

```
┌────────────────────────────────────────────────────────────────┐
│                            DISCO                               │
├────────┬────────────┬───────────────┬────────────┬─────────────┤
│  MBR   │ Partición 1│  Partición 2  │ Partición 3│ Partición 4 │
│(92bytes)│   (EXT2)  │  (Sin formato)│  (EXT2)    │ (Espacio    │
│        │            │               │            │  libre)     │
└────────┴────────────┴───────────────┴────────────┴─────────────┘
```

## Estructura detallada de una partición formateada con EXT2

Ahora, veamos en detalle cómo se organiza una partición formateada con EXT2:

```
┌─────────────────────────────────────── Partición con EXT2 ────────────────────────────────────────┐
│                                                                                                   │
│  ┌────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────────────┐  ┌───────────────────┐  │
│  │ Superbloque│  │ Bitmap      │  │ Bitmap      │  │     Tabla de       │  │    Tabla de       │  │
│  │ (92 bytes) │  │ de Inodos   │  │ de Bloques  │  │     Inodos         │  │    Bloques        │  │
│  │            │  │ (n bytes)   │  │ (3n bytes)  │  │   (n * 124 bytes)  │  │  (3n * 64 bytes)  │  │
│  └────────────┘  └─────────────┘  └─────────────┘  └────────────────────┘  └───────────────────┘  │
│                                                                                                   │
└───────────────────────────────────────────────────────────────────────────────────────────────────┘
```

Donde:
- **n** es el número de inodos que caben en la partición (calculado en la función `Mkfs`)
- Cada inodo ocupa 124 bytes
- Cada bloque ocupa 64 bytes

## Detalle de cada estructura principal

### 1. Superbloque (92 bytes)
Es la estructura que contiene toda la información del sistema de archivos:

```
┌───────────────────── Superbloque (92 bytes) ─────────────────────┐
│ S_filesystem_type   (4 bytes) - Tipo de sistema (2 = EXT2)       │
│ S_inodes_count      (4 bytes) - Total de inodos                  │
│ S_blocks_count      (4 bytes) - Total de bloques                 │
│ S_free_blocks_count (4 bytes) - Bloques libres                   │
│ S_free_inodes_count (4 bytes) - Inodos libres                    │
│ S_mtime             (19 bytes) - Fecha de montaje                │
│ S_umtime            (19 bytes) - Fecha de desmontaje             │
│ S_mnt_count         (4 bytes) - Veces montado                    │
│ S_magic             (4 bytes) - Número mágico (0xEF53)           │
│ S_inode_size        (4 bytes) - Tamaño de inodo (124)            │
│ S_block_size        (4 bytes) - Tamaño de bloque (64)            │
│ S_first_ino         (4 bytes) - Primer inodo libre               │
│ S_first_blo         (4 bytes) - Primer bloque libre              │
│ S_bm_inode_start    (4 bytes) - Inicio bitmap inodos             │
│ S_bm_block_start    (4 bytes) - Inicio bitmap bloques            │
│ S_inode_start       (4 bytes) - Inicio tabla inodos              │
│ S_block_start       (4 bytes) - Inicio tabla bloques             │
└──────────────────────────────────────────────────────────────────┘
```

### 2. Bitmap de Inodos (n bytes)
Cada bit representa un inodo:
- 0 = libre
- 1 = ocupado

```
┌─── Bitmap Inodos (n bytes) ───┐
│ 1 1 0 0 0 0 0 0 0 0 ...       │
└───────────────────────────────┘
  ^ ^
  | └─ Inodo 1 (users.txt)
  └─── Inodo 0 (carpeta raíz)
```

### 3. Bitmap de Bloques (3n bytes)
Cada bit representa un bloque:
- 0 = libre
- 1 = ocupado

```
┌─── Bitmap Bloques (3n bytes) ─┐
│ 1 1 0 0 0 0 0 0 0 0 ...       │
└───────────────────────────────┘
  ^ ^
  | └─ Bloque 1 (contenido de users.txt)
  └─── Bloque 0 (entradas de carpeta raíz)
```

### 4. Tabla de Inodos (n * 124 bytes)
Cada inodo ocupa 124 bytes:

```
┌───────────────────── Inodo (124 bytes) ─────────────────────┐
│ I_uid      (4 bytes) - ID de usuario propietario            │
│ I_gid      (4 bytes) - ID de grupo propietario              │
│ I_size     (4 bytes) - Tamaño del archivo                   │
│ I_atime    (16 bytes) - Fecha de último acceso              │
│ I_ctime    (16 bytes) - Fecha de creación                   │
│ I_mtime    (16 bytes) - Fecha de modificación               │
│ I_block    (60 bytes) - 15 apuntadores a bloques (4c/u)     │
│ I_type     (1 byte) - Tipo (0=carpeta, 1=archivo)           │
│ I_perm     (3 bytes) - Permisos (ej: 664)                   │
└─────────────────────────────────────────────────────────────┘
```

### 5. Tabla de Bloques (3n * 64 bytes)
Hay tres tipos principales de bloques:

#### a) Bloque de Carpeta (64 bytes)
```
┌───────────────── Bloque de Carpeta (64 bytes) ─────────────────┐
│ ┌───────────────────┐ ┌───────────────────┐                    │
│ │ B_name: "."       │ │ B_name: ".."      │                    │
│ │ B_inodo: 0        │ │ B_inodo: 0        │                    │
│ └───────────────────┘ └───────────────────┘                    │
│ ┌───────────────────┐ ┌───────────────────┐                    │
│ │ B_name:"users.txt"│ │ B_name: ""        │                    │
│ │ B_inodo: 1        │ │ B_inodo: -1       │                    │
│ └───────────────────┘ └───────────────────┘                    │
└────────────────────────────────────────────────────────────────┘
```

#### b) Bloque de Archivo (64 bytes)
```
┌─────────────── Bloque de Archivo (64 bytes) ───────────────┐
│                                                            │
│ B_content: "1,G,root\n1,U,root,root,123\n"                 │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

#### c) Bloque de Apuntadores (64 bytes)
```
┌───────────── Bloque de Apuntadores (64 bytes) ─────────────┐
│ B_pointers[0]: X                                           │
│ B_pointers[1]: X                                           │
│ B_pointers[2]: X                                           │
│ ...                                                        │
│ B_pointers[15]: X                                          │
└────────────────────────────────────────────────────────────┘
```

## Ejemplo visual del sistema de archivos inicial

Después de formatear con MKFS, así es cómo se ve el sistema de archivos:

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                             SISTEMA DE ARCHIVOS EXT2                                    │
├────────────┬─────────────┬─────────────┬────────────────────┬───────────────────────────┤
│ Superbloque│ Bitmap      │ Bitmap      │     Tabla de       │       Tabla de            │
│            │ de Inodos   │ de Bloques  │     Inodos         │       Bloques             │
├────────────┼─────────────┼─────────────┼────────────────────┼───────────────────────────┤
│            │ 1 1 0 0 ... │ 1 1 0 0 ... │ ┌─────────────┐    │  ┌───────────────────┐    │
│  Contiene  │             │             │ │  Inodo 0    │    │  │     Bloque 0      │    │
│  metadata  │ (inodos     │ (bloques    │ │  (carpeta /)│    │  │  (carpeta raíz)   │    │
│  del       │  0 y 1      │  0 y 1      │ └─────────────┘    │  └───────────────────┘    │
│  sistema   │  en uso)    │  en uso)    │ ┌─────────────┐    │  ┌───────────────────┐    │
│            │             │             │ │  Inodo 1    │    │  │     Bloque 1      │    │
│            │             │             │ │ (users.txt) │    │  │ (datos users.txt) │    │
│            │             │             │ └─────────────┘    │  └───────────────────┘    │
│            │             │             │      ...           │         ...               │
└────────────┴─────────────┴─────────────┴────────────────────┴───────────────────────────┘
```

## La relación entre estructuras

Veamos cómo se conectan las estructuras para formar el sistema de archivos:

```
                                  SUPERBLOQUE
                                       │
                                       │ (contiene ubicaciones)
                                       ▼
              ┌─────────────────┬─────────────────┬─────────────────┐
              │                 │                 │                 │
              ▼                 ▼                 ▼                 ▼
        BITMAP INODOS    BITMAP BLOQUES     TABLA INODOS      TABLA BLOQUES
              │                 │                 │                 │
              │                 │                 │                 │
              │                 │                 ▼                 │
              │                 │          ┌─────────────┐          │
              │                 │          │  Inodo 0    │          │
              │                 │          │ (carpeta /) │          │
              │                 │          └─────────────┘          │
              │                 │                 │                 │
              │                 │                 │ I_block[0]=0    │
              │                 │                 │                 │
              │                 │                 │                 ▼
              │                 │                 │           ┌───────────────┐
              │                 │                 └──────────►│   Bloque 0    │
              │                 │                             │ (carpeta raíz)│
              │                 │                             └───────────────┘
              │                 │                                     │
              │                 │                                     │ B_content[2].B_inodo=1
              │                 │                                     ▼
              │                 │                             ┌───────────────┐
              │                 │                             │   Inodo 1     │
              │                 │                             │  (users.txt)  │
              │                 │                             └───────────────┘
              │                 │                                     │
              │                 │                                     │ I_block[0]=1
              │                 │                                     ▼
              │                 │                             ┌───────────────┐
              │                 │                             │   Bloque 1    │
              │                 │                             │  (users.txt)  │
              │                 │                             └───────────────┘
```

## Papel de los structs en el sistema

Los structs que defines en Go tienen dos propósitos principales:

1. **Representación en memoria**: Permiten manipular las estructuras del sistema de archivos desde tu programa.



El sistema EXT2 es un sistema de archivos estructurado que almacena tanto los datos como los metadatos en un formato específico dentro de una partición. Los structs que defines en Go sirven tanto para representar esas estructuras en memoria (para que tu programa pueda manipularlas) como para definir exactamente cómo se almacenan en el disco.

Cuando formateas una partición con MKFS, lo que haces es escribir todas estas estructuras en posiciones específicas dentro de la partición, creando un sistema de archivos vacío con solo la carpeta raíz y el archivo users.txt inicial.