package initchecker

import "fmt"

func CheckInit(pairs ...any) {
	if len(pairs)%2 != 0 {
		panic("CheckInit: odd number of arguments")
	}
	for i := 0; i < len(pairs); i += 2 {
		name, ok := pairs[i].(string)
		if !ok {
			panic("CheckInit: first argument of pair must be string")
		}
		value := pairs[i+1]
		if value == nil {
			panic(fmt.Sprintf("%s dependency not initialized", name))
		}
	}
}
