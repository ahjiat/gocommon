package node
import(
	"encoding/json"
)

func New(ref interface{}) *Object {
	return &Object{Ref:ref}
}

type Object struct {
	Ref interface{}
}

func (self *Object) AddOrGet(f func(interface{}) interface{}) *Object {
	return &Object{Ref:f(self.Ref)}
}

func (self *Object) Output() interface{} {
	return self.Ref
}

func (self *Object) OutputJson() string {
	strB, _ := json.Marshal(self.Ref)
	return string(strB)
}

