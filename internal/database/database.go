package database

/*
import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	client *sql.DB
}

var volumesTable = `CREATE TABLE 'volumes' (
	'id' INTEGER PRIMARY KEY,
	'name' VARCHAR(64) NULL
)`

func InitDB(dbPath string) (db *Database, err error) {
	db = &Database{}

	db.client, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		err = fmt.Errorf("failed to open database: %s", err)
		return
	}

	err = db.createVolumesTable()
	if err != nil {
		err = fmt.Errorf("InitDB() - %s", err)
		return
	}
	return
}

func (db *Database) Close() {
	db.client.Close()
	return
}

func (db *Database) createVolumesTable() (err error) {
	volumesTable := `CREATE TABLE 'volumes' (
		'id' INTEGER PRIMARY KEY,
		'name' VARCHAR(64) NULL,
		'status' VARCHAR(255) NULL
	)`
	exists, err := db.tableExists("volumes")
	if err != nil {
		err = fmt.Errorf("createVolumesTable() - %s", err)
		return
	}
	if !exists {
		err = db.createTable(volumesTable)
		if err != nil {
			err = fmt.Errorf("createVolumesTable() - %s", err)
			return
		}
	}
	return
}

func (db *Database) tableExists(table string) (ok bool, err error) {
	stmt, err := db.client.Prepare("SELECT name FROM sqlite_master WHERE type='table' AND name=?")
	if err != nil {
		err = fmt.Errorf("tableExists() - %s", err)
		return false, err
	}
	rows, err := stmt.Query(table)
	if err != nil {
		err = fmt.Errorf("tableExists() - %s", err)
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func (db *Database) createTable(query string) (err error) {
	_, err = db.client.Exec(query)
	if err != nil {
		err = fmt.Errorf("createTable() - %s", err)
		return
	}
	return
}

func (db *Database) insertVolume(volumeID, reason, source string) (err error) {
	if reason == "" {
		status = "backed up"
	} else {
		status = fmt.Sprintf("%s (%s)", reason, source)
	}

	rows, err := db.client.Query("SELECT * FROM volumes WHERE id=?")
	if err != nil {
		err = fmt.Errorf("insertVolume() - %s", err)
		return err
	}
	defer rows.Close()
	if rows.Next() {
		stmt, err := db.client.Prepare("UPDATE volumes SET name = ?, status = ? WHERE id = ?")
		if err != nil {
			err = fmt.Errorf("insertVolume() - %s", err)
			return err
		}
		_, err := stmt.Query(volumeName, status, volumeID)
		if err != nil {
			err = fmt.Errorf("insertVolume() - %s", err)
			return err
		}
	}
}
*/
