(module
  (import "wasi_snapshot_preview1" "proc_exit" (func $exit (param i32)))
  (memory (export "memory") 1 1 shared)
  (func (export "_start") (call $exit (i32.atomic.rmw.add (i32.const 0) (i32.const 5)) ) ))
