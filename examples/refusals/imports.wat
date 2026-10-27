(module
  (import "env" "log" (func $log (param i32)))
  (func (export "go") (call $log (i32.const 1))))
