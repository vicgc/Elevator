package elevator

const DB_STATUS_MOUNTED = 1
const DB_STATUS_UNMOUNTED = 0

const (
	DB_GET 		= "GET"
	DB_PUT 		= "PUT"
	DB_DELETE 	= "DELETE"
	DB_RANGE 	= "RANGE"
	DB_SLICE 	= "SLICE"
	DB_BATCH 	= "BATCH"
	DB_MGET 	= "MGET"
	DB_PING 	= "PING"
	DB_CONNECT 	= "DBCONNECT"
	DB_MOUNT 	= "DBMOUNT"
	DB_UMOUNT 	= "DBUMOUNT"
	DB_CREATE 	= "DBCREATE"
	DB_DROP 	= "DBDROP"
	DB_LIST 	= "DBLIST"
	DB_REPAIR 	= "DBREPAIR"
)