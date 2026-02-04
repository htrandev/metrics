package p1

import "log"

func fatalFunc() {
	log.Fatal("fatal") // want "calling log.Fatal outside main"
}
