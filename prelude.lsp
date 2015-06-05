(macro defn (name args body)
       (def name
            (lambda args body)))
