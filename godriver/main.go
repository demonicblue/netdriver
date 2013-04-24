package ivbsnetd

import(
	"fmt"
	"flag"
	"syscall"
)

const DATASIZE = 1024*1024*50

func main() {
	make([]uint8, DATASIZE)
	
	var ndb_path string
	flar.StrinVar(&nbd_path, "NBDPath", "Path to NBD device")
	
	fd := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	
	
	
	
}

