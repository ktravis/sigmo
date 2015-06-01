(defn greet (name)
      (print "hi" name))

(def world "world")

(greet world)

(defn yell (msg) (cat msg "!"))

(print (yell "HIII"))
(print (cat "a" "b" 1 2 "lolol" 1.23))

(def how-lumpy 42)

(if (lte how-lumpy 10)
  (print "not very...")
  (print "SO FREAKIN LUMPY"))

(def a 10)
(def ls (1 2 3 4 5 6))

; (for (identifier list) stuff)
(def squares
  (for (a (cons 7 ls))
    (* a a)))

; (let (name val name2 val2...) stuff)
(let (a) (print a)) ; nil

(let (a 10 b (+ a 2))
  (print a b)) ; 10 12

(print a) ; 10

(print squares)
