package webgo
import "fmt"

type Auth struct {}

func (a *Auth) Run () {
	fmt.Println("AUTH")
}


func init(){

}