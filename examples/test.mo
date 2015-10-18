(import "examples/prelude.mo")
(defn fib (n)
  (if (> n 2)
    (+ (fib (- n 1)) (fib (- n 2)))
    1))

(println (for (a (1 2 3 4 5 6 7 8 9 10)) (fib a)))
