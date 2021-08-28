package serialize
import(
	"encoding/json"
	"unsafe"
)

func Unmarshal(data []byte, val interface{}){
	json.Unmarshal(data,val)
}

func Marshal(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

func BytesToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}
