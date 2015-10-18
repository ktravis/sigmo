(macro quote (ls)
  '(ls...))
(macro defn (name args body)
       (def name
            (lambda args body)))

(defn min (a b)
      (if (< a b) a b))

(defn max (a b)
      (if (> a b) a b))

(defn nil? (a) (= a nil))

(defn empty? (a) (= (len a) 0))

(defn map (fn ls)
      (for (a ls) (fn a)))

(defn reduce (fn ls acc)
      (let (x acc)
        (for (a ls)
          (set! x (fn x a)))
        x))

(defn sum (ls) (reduce + ls 0))

(defn filter (fn ls)
      (let (out ())
        (for (a ls)
          (if (fn a)
            (set! out (cons out a))))
        out))

(defn int? (a)
  (= (type a) #int))

(defn float? (a)
  (= (type a) #float))

(defn bool? (a)
  (= (type a) #bool))

(defn type? (a)
  (= (type a) #type))

(defn string? (a)
  (= (type a) #string))

(defn list? (a)
  (= (type a) #list))

(defn hash? (a)
  (= (type a) #hash))
