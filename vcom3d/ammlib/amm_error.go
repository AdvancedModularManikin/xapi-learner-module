
package ammlib

type Error struct {
   code int
   msg string
}

func error0 () (Error) {
   return Error{code: 0}
}

func error1 () (Error) {
   return Error{code: 1, msg: "Func name already exists."}
}

func error2 () (Error) {
   return Error{code: 2, msg: "Func name does not exist."}
}
