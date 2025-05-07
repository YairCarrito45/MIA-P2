package Estructuras

type UserInfo struct {
	Id     string // el id de la particion en donde se abrío secion el usuario
	IdGrp  int32  //id del grupo al que pertenece el usuario
	IdUsr  int32  //id del usuario
	Nombre string //saber que usuario es (identifica si es root o cualquir otro)
	Status bool   //si esta iniciada la sesion
	PathD  string //path del disco
}

var UsuarioActual UserInfo
