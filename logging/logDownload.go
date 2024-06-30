package logging

import (
	"database/sql"
	"regexp"
	"strings"

	_ "modernc.org/sqlite" // Importing SQLite driver for database interaction.
)

var Dbsql *sql.DB

func InitSql(data string) *sql.DB {
	db, _ := sql.Open("sqlite", "file:"+data)
	return db
}
func StartSQL(data string){
	Dbsql=InitSql(data)
}
func InsertDownload(data ...string) {
	if Dbsql != nil{
		splitter := regexp.MustCompile(`\.`)
		insertion:= splitter.Split(data[0],-1)
		pkg:=GetPackage(insertion)
		ver:= GetVersion(insertion)
		timeus:=data[1]
		Dbsql.Exec("INSERT INTO download(program,version,time) VALUES (?,?,?)", pkg, ver,timeus )
	}
}
func GetPackage(file []string) string{
	data := strings.Split(file[0], "/")
	return data[len(data)-1]+"."+file[1]
}
func GetVersion(file []string) string{
	matcher:=regexp.MustCompile(`\d+`)
	ndt:=[]string{}
	for _,i := range file{
		if matcher.MatchString(i){
			ndt = append(ndt, i)
		}
	}
	return strings.Join(ndt, ".")
}
