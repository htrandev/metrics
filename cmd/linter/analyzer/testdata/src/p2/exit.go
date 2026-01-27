package p2

import "os"

func exitFunc() {
	os.Exit(1) // want "calling os.Exit outside main"
}
