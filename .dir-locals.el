;;; Directory Local Variables
;;; For more information see (info "(emacs) Directory Variables")
((go-mode
  . ((go-test-args . "-tags libsqlite3 -timeout 1m")
     (eval
      . (set
	 (make-local-variable 'flycheck-go-build-tags)
	 '("libsqlite3"))))))
