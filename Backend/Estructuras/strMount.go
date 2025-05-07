package Estructuras

/*
Almacena la informacion de los Discos montados:
Se asigna una letra a cada disco montado y
va sumando 1 cada vez que se monta otra particion en dicho disco
*/
var Pmontaje []Mount

type Mount struct {
	MPath  string //Path del Disco
	Letter byte   //Letra que se le asigna
	Cont   int    //COntador del numero de particion montada
}

// Para llevar el conteo de los discos
func AddPathM(path string, L byte, cont int) {
	Pmontaje = append(Pmontaje, Mount{MPath: path, Letter: L, Cont: cont})
}

// ==============================================================================

// esta informacion si se borra, pero realmente sigue escrita en los disco.
// Almacena la informacion de cada Id junto a su Path
var Montadas []mountAlready

/*

Montadas = [
    { Id: "291A", PathM: "/dev/sda" },
    { Id: "292A", PathM: "/dev/sda" },
    { Id: "291B", PathM: "/dev/sdb" }

*/

type mountAlready struct {
	Id    string //Id de la particion
	PathM string //Path del disco al que pertenece la particion
}

// Ingresar particion montada
func AddMontadas(id string, path string) {
	Montadas = append(Montadas, mountAlready{Id: id, PathM: path})
}
