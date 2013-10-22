package elevator

import (
	"bytes"
	leveldb "github.com/jmhodges/levigo"
	"strconv"
)

var database_commands = map[string]func(*Database, *Message) (*Response, error){
	DB_GET:    Get,
	DB_MGET:   MGet,
	DB_PUT:    Put,
	DB_DELETE: Delete,
	DB_RANGE:  Range,
	DB_SLICE:  Slice,
	DB_BATCH:  Batch,
}

var store_commands = map[string]func(*DatabaseRegistry, *Message) (*Response, error){
	DB_CREATE:  DatabaseCreate,
	DB_DROP:    DatabaseDrop,
	DB_CONNECT: DatabaseConnect,
	DB_MOUNT:   DatabaseMount,
	DB_UMOUNT:  DatabaseUnmount,
	DB_LIST:    DatabaseList,
}

func Get(db *Database, message *Message) (*Response, error) {
	var response *Response
	var key string = message.Args[0]

	ro := leveldb.NewReadOptions()
	value, err := db.Connector.Get(ro, []byte(key))
	if err != nil {
		response = NewFailureResponse(KEY_ERROR, err.Error())
	} else {
		response = NewSuccessResponse([]string{string(value)})
	}

	return response, nil
}

func Put(db *Database, message *Message) (*Response, error) {
	var response *Response
	var key string = message.Args[0]
	var value string = message.Args[1]

	wo := leveldb.NewWriteOptions()
	err := db.Connector.Put(wo, []byte(key), []byte(value))
	if err != nil {
		response = NewFailureResponse(VALUE_ERROR, string(err.Error()))
	} else {
		response = NewSuccessResponse([]string{})
	}

	return response, nil
}

func Delete(db *Database, message *Message) (*Response, error) {
	var response *Response
	var key string = message.Args[0]

	wo := leveldb.NewWriteOptions()
	err := db.Connector.Delete(wo, []byte(key))
	if err != nil {
		response = NewFailureResponse(KEY_ERROR, string(err.Error()))
	} else {
		response = NewSuccessResponse([]string{})
	}

	return response, nil
}

func MGet(db *Database, message *Message) (*Response, error) {
	var response *Response
	var data []string = make([]string, len(message.Args))

	readOptions := leveldb.NewReadOptions()
	snapshot := db.Connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)

	if len(message.Args) > 0 {
		start := message.Args[0]
		end := message.Args[len(message.Args)-1]

		keysIndex := make(map[string]int)

		for index, element := range message.Args {
			keysIndex[element] = index
		}

		it := db.Connector.NewIterator(readOptions)
		defer it.Close()
		it.Seek([]byte(start))

		for ; it.Valid(); it.Next() {
			if bytes.Compare(it.Key(), []byte(end)) > 1 {
				break
			}

			if index, present := keysIndex[string(it.Key())]; present {
				data[index] = string(it.Value())
			}
		}

	}

	response = NewSuccessResponse(data)
	forwardResponse(response, message)

	db.Connector.ReleaseSnapshot(snapshot)

	return response, nil
}

func Range(db *Database, message *Message) (*Response, error) {
	var response *Response
	var data []string
	var start string = message.Args[0]
	var end string = message.Args[1]

	readOptions := leveldb.NewReadOptions()
	snapshot := db.Connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)

	it := db.Connector.NewIterator(readOptions)
	defer it.Close()
	it.Seek([]byte(start))

	for ; it.Valid(); it.Next() {
		if bytes.Compare(it.Key(), []byte(end)) >= 1 {
			break
		}
		data = append(data, string(it.Key()), string(it.Value()))
	}

	response = NewSuccessResponse(data)
	forwardResponse(response, message)

	db.Connector.ReleaseSnapshot(snapshot)

	return response, nil
}

func Slice(db *Database, message *Message) (*Response, error) {
	var response *Response
	var data []string
	var start string = message.Args[0]

	limit, _ := strconv.Atoi(message.Args[1])
	readOptions := leveldb.NewReadOptions()
	snapshot := db.Connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)

	it := db.Connector.NewIterator(readOptions)
	defer it.Close()
	it.Seek([]byte(start))

	i := 0
	for ; it.Valid(); it.Next() {
		if i >= limit {
			break
		}

		data = append(data, string(it.Key()), string(it.Value()))
		i++
	}

	response = NewSuccessResponse(data)
	forwardResponse(response, message)

	db.Connector.ReleaseSnapshot(snapshot)

	return response, nil
}

func Batch(db *Database, message *Message) (*Response, error) {
	var response *Response
	var operations *BatchOperations
	var batch *leveldb.WriteBatch = leveldb.NewWriteBatch()

	operations = BatchOperationsFromMessageArgs(message.Args)

	for _, operation := range *operations {
		switch operation.OpCode {
		case SIGNAL_BATCH_PUT:
			batch.Put([]byte(operation.OpArgs[0]), []byte(operation.OpArgs[1]))
		case SIGNAL_BATCH_DELETE:
			batch.Delete([]byte(operation.OpArgs[0]))
		}
	}

	wo := leveldb.NewWriteOptions()
	err := db.Connector.Write(wo, batch)
	if err != nil {
		response = NewFailureResponse(VALUE_ERROR, string(err.Error()))
	} else {
		response = NewSuccessResponse([]string{})
	}

	return response, nil
}

func DatabaseCreate(db_store *DatabaseRegistry, message *Message) (*Response, error) {
	var response *Response
	var dbName string = message.Args[0]

	err := db_store.Add(dbName)
	if err != nil {
		response = NewFailureResponse(DATABASE_ERROR, string(err.Error()))
	} else {
		response = NewSuccessResponse([]string{})
	}

	return response, nil
}

func DatabaseDrop(db_store *DatabaseRegistry, message *Message) (*Response, error) {
	var response *Response
	var dbName string = message.Args[0]

	err := db_store.Drop(dbName)
	if err != nil {
		response = NewFailureResponse(DATABASE_ERROR, string(err.Error()))
	} else {
		response = NewSuccessResponse([]string{})
	}

	return response, nil
}

func DatabaseConnect(db_store *DatabaseRegistry, message *Message) (*Response, error) {
	var response *Response
	var dbName string = message.Args[0]

	dbUid, exists := db_store.NameToUid[dbName]

	if exists {
		response = NewSuccessResponse([]string{dbUid})
	} else {
		response = NewFailureResponse(DATABASE_ERROR, "Database does not exist")
	}

	return response, nil
}

func DatabaseList(db_store *DatabaseRegistry, message *Message) (*Response, error) {
	var response *Response

	dbNames := db_store.List()
	data := make([]string, len(dbNames))

	for index, dbName := range dbNames {
		data[index] = dbName
	}

	response = NewSuccessResponse(data)
	return response, nil
}

func DatabaseMount(db_store *DatabaseRegistry, message *Message) (*Response, error) {
	var response *Response
	var dbName string = message.Args[0]

	dbUid, exists := db_store.NameToUid[dbName]

	if exists {
		err := db_store.Mount(db_store.Container[dbUid].Name)
		if err != nil {
			return nil, err
		}

		response = NewSuccessResponse([]string{})
	} else {
		response = NewFailureResponse(DATABASE_ERROR, "Database does not exist")
	}

	return response, nil

}

func DatabaseUnmount(db_store *DatabaseRegistry, message *Message) (*Response, error) {
	var response *Response
	var dbName string = message.Args[0]

	dbUid, exists := db_store.NameToUid[dbName]

	if exists {
		err := db_store.Unmount(dbUid)
		if err != nil {
			return nil, err
		}

		response = NewSuccessResponse([]string{})
	} else {
		response = NewFailureResponse(DATABASE_ERROR, "Database does not exist")
	}

	return response, nil
}
