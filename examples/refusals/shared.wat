(module
  (memory (export "mem") 1 1 shared)
  (func (export "wait") (result i32)
    (memory.atomic.wait32 (i32.const 0) (i32.atomic.load (i32.const 0)) (i64.const 1000000)))
  (func (export "inc") (result i32)
    (i32.atomic.rmw.add (i32.const 0) (i32.const 1))))
