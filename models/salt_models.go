package models

import (
	"errors"
	"database/sql"
	"fmt"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
)
type Fields    map[string]Field
type Objects   []Object
type Object    map[string]interface{}

type Field struct {
	Type            Type
	AutoIncrement   bool
	NotNull         bool
}

type Model struct {
	Name             string
	Fields           Fields
	Objects          Objects
	PrimaryKey       string
}

type Models []Model

type Type string

type Database struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Database string `json:"Database"`
}

const(
	CharField     Type = "string"
	Integer       Type = "int"
	Float         Type = "float"
	Boolean       Type = "bool"
)
var errMigrated error = errors.New("Error : There is already a table name with the same name as the model")

var modelstore Models
var database   Database

func (models *Models) Register() (error) {
	for _,val := range *models {
		err := val.Register()
		if  err != nil {
			return err
		}
	}
	return nil
}

func (model *Model) IsMigrated()(bool) {
	if model.Check() == errMigrated {
		return true
	}
	return false
}

func (model *Model) Register() (error) {
	if (len(modelstore) == 0) {
		modelstore = make(Models,1)
	}
	for _,val := range modelstore {
		if (val.Name == model.Name) {
			return errors.New("Error : There is another model with the same name {"+val.Name+"} already registered.")
		}
	}
	if _,ok := model.Fields[model.PrimaryKey]; ( !ok ) {
		fmt.Println(ok,model.PrimaryKey,model.Fields[model.PrimaryKey])
		return errors.New("Error : Specified Primary key is not defined in the field list")
	}
	err := model.Check()
	if err != nil {
		return err
	}
	err = model.AddToDataBase()
	if err != nil {
		return err
	}
	modelstore = append(modelstore,*model)
	println("Registered model : " + model.Name)
	return nil
}

//
func (model *Model) Check () (error) {
	db, err := sql.Open("mysql", database.Username+":"+database.Password+"@/"+database.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var temp string
		err := rows.Scan(&temp)
		if err != nil {
			return err
		}
		if model.Name == temp {
			return errMigrated
		}
	}
	return nil
}

//SetDatabaseConfig : Set Database configuration from the Parent module "salt"
//This step is done when the app.Configure is called.
func SetDatabaseConfig(datab Database) (error) {
	db, err := sql.Open("mysql", datab.Username+":"+datab.Password+"@/"+datab.Database)
	defer db.Close()
	if (err == nil) {
		database = datab
		fmt.Println("Able to open Connection to database")
	}
	return err
}

//AddToDatabase : Create a database for the corresponding Model
func (model *Model) AddToDataBase() (error) {
	db, err := sql.Open("mysql", database.Username+":"+database.Password+"@/"+database.Database)
	defer db.Close()
	if err != nil {
		return nil
	}
	query := "CREATE TABLE `"
	query += model.Name
	query += "` ( "
	for fieldName,fieldType  := range model.Fields {
		query += fieldName + " "
		switch fieldType.Type {
		case CharField:
			query += "TEXT"
		case Integer:
			query += "INT"
		case Float:
			query += "DOUBLE"
		case Boolean:
			query += "BOOL"
		}
		if fieldType.AutoIncrement {
			query += " AUTO_INCREMENT"
		}
		if fieldType.NotNull {
			query +=  " NOT NULL"
		}
		query += ","
	}
	if ( model.PrimaryKey != "" ) {
		query += "PRIMARY KEY(" + model.PrimaryKey + ")"
	} else {
		query = query[:len(query)-1]
	}
	query += ")"
	fmt.Println(query)
	_, err = db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

//
func (model *Model)GetAll()(error)  {
	db, err := sql.Open("mysql", database.Username+":"+database.Password+"@/"+database.Database)
	defer db.Close()
	if err != nil {
		return nil
	}
	rows, err := db.Query("SELECT * FROM `"+model.Name+"`")
	defer rows.Close()
	if err != nil {
		return err
	}
	arr , err := rows.Columns()
	if err != nil {
		return err
	}
	inner := make([]interface{},len(arr))
	elements := make([]interface{},len(arr))
	for i,_ := range inner {
		elements[i] = &inner[i]
	}
	for rows.Next() {
		err = rows.Scan(elements...)
		for i,val := range arr {
			tstr := fmt.Sprintf("%#s",(*elements[i].(*interface{})).([]uint8))
			fmt.Println(val,tstr)
			tobj := make(Object,0)
			switch model.Fields[val].Type {
			case CharField:
				tobj[val] = tstr
			case Integer:
				tobj[val],err = strconv.Atoi(tstr)
				if err != nil {
					return err
				}
			case Float:
				tobj[val],err = strconv.ParseFloat(tstr, 64)
				if err != nil {
					return err
				}
			case Boolean:
				tobj[val],err = strconv.ParseBool(tstr)
				if err != nil {
					return err
				}
			}
			model.Objects = append(model.Objects,tobj)
		}
	}
	return nil
}

//
func (model *Model) AddNewRecord (object Object) (error) {
	db, err := sql.Open("mysql", database.Username+":"+database.Password+"@/"+database.Database)
	defer db.Close()
	if err != nil {
		return nil
	}
	rows, err := db.Query("SELECT * FROM `"+model.Name+"`")
	defer rows.Close()
	if err != nil {
		return err
	}
	arr , err := rows.Columns()
	if err != nil {
		return err
	}
	var query string = "INSERT INTO `" + model.Name + "`"
	var columns string = " ("
	var values string = " VALUES ("
	for _,val := range arr {
		if value,ok := object[val] ; ok {
			columns += "`" + val + "`,"
			switch model.Fields[val].Type {
			case CharField:
				values += "\""+value.(string)+"\","
			case Integer:
				values += fmt.Sprint(value.(int)) + ","
			case Float:
				values += fmt.Sprint(value.(float64)) + ","
			case Boolean:
				if value.(bool) {
					values += "true,"
				} else {
					values += "false,"
				}
			}
		}
	}
	columns = columns[:len(columns)-1]+")"
	values = values[:len(values)-1]+")"
	_, err = db.Query(query+columns+values)
	if err != nil {
		return err
	}
	return err
}
