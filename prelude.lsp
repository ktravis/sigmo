(macro defn (name args body)
       (def name
            (lambda args body)))

(defn min (a b)
      (if (< a b) a b))

(defn max (a b)
      (if (> a b) a b))

(defn *max (ls)
      (reduce max ls (head ls)))

(defn *min (ls)
      (reduce min ls (head ls)))

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
