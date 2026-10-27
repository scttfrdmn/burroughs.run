(module
  (func $f (export "f") (param i32) (result i32)
    (if (result i32) (i32.eqz (local.get 0))
      (then (i32.const 0))
      (else (return_call $f (i32.sub (local.get 0) (i32.const 1)))))))
