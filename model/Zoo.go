package model

type Zoo struct {
	Animals []*Animal
	Manager string
}
type Animal struct{
	Color string
	Height int
}
func BuildZoo(manager string) *Zoo{
	return &Zoo{
		Animals: BuildAnimal("red", 100,2),
		Manager: manager,
	}
}
func BuildAnimal(color string,height int,len int) []*Animal{
	var animalList = []*Animal{}
	for i:=0;i<len;i++ {
		animalList = append(animalList, &Animal{Color: color,Height: height+i})
	}
	return animalList
}
