; a simple sigmo repl, implemented in sigmo
(let (looping true line "")
  (while looping
    (print "> ")
    (set! line (trim (input) "\n"))
    (if (eq line "exit")
      (set! looping false)
      (do
        (def last (eval line))
        (if (neq nil last)
          (println (eval line)))))))

