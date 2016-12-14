(import "prelude.mo")

(macro pop! (ls)
 (let (temp (head ls))
  (set! ls (tail ls))
  temp))

(defn stack-op (op argc stack) (do
  (let (temp () count 0)
    (while (< count argc)
     (set! temp (cons (pop! stack) temp))
     (set! count (+ 1 count)))
    (cons (op temp...) temp))))

(def dict {
 "." (lambda (s) (stack-op print 1 s))
 "CR" (lambda (s) (do (println) s))
 "+" (lambda (s) (stack-op + 2 s))
 "-" (lambda (s) (stack-op - 2 s))
 "*" (lambda (s) (stack-op * 2 s))
 "/" (lambda (s) (stack-op / 2 s))
 ":" (lambda (s) (do (set! new-word true) (cons ":" s)))
 ";" (lambda (s) (do
     (let (temp () x nil)
      (while (neq x ":")
       (set! x (pop! s))
       (set! temp (cons x temp)))
      (set! temp (tail temp))
      (hset! dict (head temp) (lambda (s) (feval (cons (rev (tail temp)) s)))))
     (set! new-word false)
     s))
 })

(defn feval (s)
  (let (tok (head s))
   (if (or (eq tok ";") (and (not new-word) (hcontains dict tok)))
    (set! s ((hget dict (pop! s)) s)))
  s))
  

(defn forth (proc)
 (let (stack () new-word false)
  (for (tok (split proc " ")) (do
   (guard (set! tok (parse-int tok)))
   (set! stack (feval (cons tok stack)))))
  stack))

(forth ": INC 1 + ; 1 INC . CR")
