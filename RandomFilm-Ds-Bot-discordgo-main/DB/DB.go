package DB

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func RandomFilm() (name, url string) {

	psqlconn := fmt.Sprintf("host= %s port = %d user = %s password = %s dbname = %s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	rows, err := db.Query(`select id, name, url from "UsersFilms" where id = (SELECT floor(random()*((select count(*) from "UsersFilms")-1+1))+1);`)
	if err != nil {
		log.Fatal(err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var id int

	for rows.Next() {
		err := rows.Scan(&id, &name, &url)
		if err != nil {
			log.Println(err)
			return
		}
	}

	_, e := db.Exec(`DELETE FROM "UsersFilms" WHERE id = $1;`, id)
	if e != nil {
		log.Fatal(err)
	}

	_, e = db.Exec(`update "UsersFilms" set id = id -1 where id > $1;`, id)
	if e != nil {
		log.Fatal(err)
	}

	return name, url
}

func ListFilms() (FilmDict map[string]string) {
	psqlconn := fmt.Sprintf("host= %s port = %d user = %s password = %s dbname = %s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	rows, err := db.Query(`select name, url from "UsersFilms";`)
	if err != nil {
		log.Fatal(err)
	}

	var (
		name string
		url  string
	)

	FilmDict = make(map[string]string)

	for rows.Next() {
		err := rows.Scan(&name, &url)
		FilmDict[name] = url
		if err != nil {
			log.Println(err)
			return
		}
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	return FilmDict
}

func AddFilm(name, url string) (answer string) {
	psqlconn := fmt.Sprintf("host= %s port = %d user = %s password = %s dbname = %s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if url != "" {
		res, err := db.Query(`SELECT * FROM addwhithurl($1, $2)`, name, url)
		if err != nil {
			answer = "Такой фильм уже есть"
			return answer
		}
		answer = "done"
		err = res.Close()
		if err != nil {
			return err.Error()
		}
	} else {

		res, err := db.Query(`SELECT * FROM addwhithouturl($1)`, name)
		if err != nil {
			answer = "Такой фильм уже есть"
			return answer
		}
		answer = "done"
		err = res.Close()
		if err != nil {
			return err.Error()
		}
	}

	return answer
}
