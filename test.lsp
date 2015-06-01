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
