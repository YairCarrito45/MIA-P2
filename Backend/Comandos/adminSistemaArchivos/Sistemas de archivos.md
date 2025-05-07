# administrción del sistema de archivos

Administración del sistema de archivos: los comandos de este apartado simularan:
		 el formateo de las particiosn
		 administracion de usuarios, carpetas y archivos 
		 -> Debe existir una sesion activa, a exepcion del mkfs y el login


# Estructura General de una Partición EXT2

Una partición EXT2 se organiza en cinco secciones principales:

```
| Superbloque | Bitmap Inodos | Bitmap Bloques | Inodos | Bloques |
```

Cada una cumple una función específica dentro del sistema de archivos.

## 1. Superbloque

El superbloque es como la "tabla de contenidos" del sistema de archivos. Contiene información esencial sobre la estructura de la partición.

### ¿Qué almacena?
- Número total de inodos y bloques
- Cantidad de inodos y bloques libres
- Fecha y hora del último montaje y desmontaje
- Tamaños de inodos y bloques
- Ubicaciones de las diferentes secciones (bitmaps, inodos, bloques)
- Un número mágico (0xEF53) que identifica al sistema como EXT2

**Ejemplo intuitivo:** Es como la primera página de un libro, donde está el índice y la información general sobre el contenido.

## 2. Bitmap de Inodos

Es un mapa de bits que indica cuáles inodos están ocupados y cuáles están libres.

### ¿Qué almacena?
- Cada bit representa un inodo:
  - `0`: Inodo libre
  - `1`: Inodo en uso

**Ejemplo intuitivo:** Piensa en un estacionamiento con 10 espacios. Un bitmap sería una lista donde los espacios ocupados se marcan con `1` y los libres con `0`:  
`[1,1,0,1,0,0,0,1,1,0]`

## 3. Bitmap de Bloques

Similar al bitmap de inodos, pero en lugar de inodos, rastrea bloques de datos.

### ¿Qué almacena?
- Cada bit indica el estado de un bloque:
  - `0`: Bloque libre
  - `1`: Bloque ocupado

**Ejemplo intuitivo:** Si tienes una libreta con 20 páginas y quieres saber cuáles están usadas, podrías hacer una lista de 20 bits donde cada bit representa una página.

## 4. Inodos (Index Nodes)

Los inodos son estructuras que contienen la información sobre cada archivo o carpeta del sistema.

### ¿Qué almacena cada inodo?
- Identificadores de usuario y grupo (UID y GID)
- Tamaño del archivo
- Fechas (creación, modificación, último acceso)
- Permisos de acceso
- Tipo de archivo (archivo regular o carpeta)
- Punteros a los bloques de datos

**Ejemplo intuitivo:** Un inodo es como la ficha bibliográfica de un libro en una biblioteca. No contiene el libro en sí, sino la información sobre él y dónde encontrarlo.

## 5. Bloques

Los bloques son las unidades donde se almacena el contenido real de los archivos o carpetas.

### Tipos de bloques en EXT2
- **Bloques de carpetas**: Almacenan nombres de archivos y la referencia a su inodo.
- **Bloques de archivos**: Contienen los datos reales de un archivo.
- **Bloques de apuntadores**: Se usan para acceder a más bloques de datos en archivos grandes.

**Ejemplo intuitivo:**  
- Si un inodo es la ficha bibliográfica, los bloques son las páginas del libro.  
- Un bloque de carpeta es como una lista con nombres de archivos.  
- Un bloque de archivo es el contenido real (como texto o imágenes).  
- Un bloque de apuntadores es como un índice adicional que te dice "para seguir leyendo, ve a estas otras páginas".

## ¿Cómo funciona la relación entre estos elementos?

1. **Cuando accedes a un archivo:**
   - El sistema busca el inodo correspondiente.
   - El inodo tiene punteros a los bloques donde está el contenido.
   - El sistema lee esos bloques para mostrar el contenido.

2. **Cuando creas un archivo nuevo:**
   - Se busca un inodo libre en el bitmap de inodos.
   - Se buscan bloques libres en el bitmap de bloques.
   - Se actualiza el inodo con los punteros a esos bloques.
   - Se actualiza el superbloque con la nueva información.

## Cálculo de la estructura del sistema EXT2

Para determinar cuántos inodos y bloques se pueden crear en una partición, se usa la siguiente fórmula:

```
tamaño_particion = sizeOf(superblock) + n + 3*n + n*sizeOf(inodos) + 3*n*sizeOf(block)
numero_estructuras = floor(n)
```

### Relación entre inodos y bloques
Por cada inodo (`n`), hay **tres** bloques (`3*n`). Esto significa que en una partición EXT2, los inodos y los bloques siguen una proporción de **1:3**.

---

Con esta explicación, ahora puedes entender mejor cómo funciona el sistema de archivos EXT2 y cómo implementar un comando **MKFS** para formatearlo correctamente.