package main

import (
	//Import the models package that we just created. You need to prefix this with
	// whatever module path you set up back in chapter 02.01 (Project Setup and Creating
	// a Module) so that the import statement looks like this:
	// "{your-module-path}/internal/models". If you can't remember what module path you
	// used, you can find it at the top of the go.mod file.
	"LetsGo/internal/models"
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include fields for the two custom loggers, but
// we'll add more to it as the build progresses
// Add a snippets field to the application struct. This will allow us to
// make the SnippetModel object available to our handlers.
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *models.SnippetModel
}

func main() {
	//Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the
	// flag will be stored in the addr variable at runtime
	addr := flag.String("addr", ":4000", "http listen address")
	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "/snippetbox?parseTime=true", "MySQL data source name")
	// Importantly, we use the flag.Parse() function to parse the command-line flag.
	//This reads in the command-line flag value and assigns it to the addr
	//variable. You need to call this *before* you use the addr variable
	//otherwise it will always contain the default value of ":4000". If any errors are
	//encountered during parsing the application will be terminated.
	flag.Parse()
	//Use log.New() to create a logger for writing information messages. This takes
	// three parameters: the destination to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "[INFO]:\t", log.Ldate|log.Ltime)
	//Create a logger for writing error messages in the same way, but use stderr as
	// the destination and use the log.Lshortfile flag to include the relevant
	// file name and line number.
	errLog := log.New(os.Stderr, "[ERROR]:\t", log.Ldate|log.Ltime|log.Llongfile)

	//To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)

	if err != nil {
		errLog.Fatal(err)
	}
	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()
	//Initialize a models.SnippetModel instance and add it to the application
	// dependencies
	app := &application{
		errorLog: errLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{DB: db},
	}
	//Initialize a new http.Server struct. We set the Addr and Handler fields so
	// that the server uses the same network address and routes as before, and set
	// the ErrorLog field so that the server now uses the custom errorLog logger in
	// the event of any problems.
	srv := &http.Server{
		Addr:     *addr,
		Handler:  app.routes(),
		ErrorLog: errLog,
	}
	infoLog.Printf("Listening on ::%v", *addr)
	//err := http.ListenAndServe(*addr, mux)
	err = srv.ListenAndServe()
	if err != nil {
		errLog.Fatal(err)
	}
}

// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
