package Estructuras

/*
----------> EBR <----------

	es un descriptor de una unidad lógica ya que es contiene la información
	y datos de la misma y apunta hacia el espació donde se escribirá el siguiente EBR

Atributos:

	part_mount  - char      - Indica si la partición está montada o no
	part_fit    - char      - Tipo de ajuste de la partición. B (Best), F(First) o W (worst)
	part_start  - int       - Indica en qué byte del disco inicia la partición
	part_s      - int       - Contiene el tamaño total de la partición en bytes.
	part_next   - int       - Byte en el que está el próximo EBR. -1 si no hay siguiente
	part_name   - char[16]  - Nombre de la partición

*/

type EBR struct {
	EbrP_mount [1]byte  //part_mount (si esta montada)
	EbrP_fit   [1]byte  //part_fit
	EbrP_start int32    //part_start
	EbrP_size  int32    //part_s
	EbrP_next  int32    //part_next
	EbrP_name  [16]byte //part_name
	EbrType    [1]byte
}
