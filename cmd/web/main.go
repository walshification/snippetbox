package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	// Import the models package that we just created. You need to prefix this with
	// whatever module path you set up back in chapter 02.01 (Project Setup and Creating
	// a Module) so that the import statement looks like this:
	// "{your-module-path}/internal/models". If you can't remember what module path you
	// used, you can find it at the top of the go.mod file.
	"snippetbox.walshification.com/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	addr      string
	staticDir string
}

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include fields for the two custom loggers, but
// we'll add more to it as the build progresses.
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	// Add a snippets field to the application struct. This will allow us to
	// make the SnippetModel object available to our handlers.
	snippets *models.SnippetModel
	// Add a templateCache field to the application struct
	templateCache map[string]*template.Template
}

func main() {
	var cfg config

	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the
	// flag will be stored in the addr variable at runtime.
	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	// Do the same for our static assets path.
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "", "MySQL data source name")

	// Importantly, we use the flag.Parse() function to parse the command-line flag.
	// This reads in the command-line flag value and assigns it to the addr
	// variable. You need to call this *before* you use the addr variable
	// otherwise it will always contain the default value of ":4000". If any errors are
	// encountered during parsing the application will be terminated.
	flag.Parse()

	// Use log.New() to create a logger for writing information messages. This takes
	// three parameters: the destination to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.LUTC)

	// Create a logger for writing error messages in the same way, but use stderr as
	// the destination and use the log.Lshortfile flag to include the relevant
	// file name and line number.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	// To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize a new instance of our application struct, containing the
	// dependencies.
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		// Initialize a models.SnippetModel instance and add it to the application
		// dependencies.
		snippets: &models.SnippetModel{DB: db},
		// Add template cache to the application dependencies.
		templateCache: templateCache,
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields so
	// that the server uses the same network address and routes as before, and set
	// the ErrorLog field so that the server now uses the custom errorLog logger in
	// the event of any problems.
	srv := &http.Server{
		Addr:     cfg.addr,
		ErrorLog: errorLog,
		// Call the new app.routes() method to get the servemux containing our routes.
		Handler: app.routes(),
	}

	// Write messages using the two new loggers, instead of the standard logger.
	infoLog.Printf("Starting server on %s", cfg.addr)

	// Call the ListenAndServe() method on our new http.Server struct.
	// Because the err variable is now already declared in the code above, we need
	// to use the assignment operator = here, instead of the := 'declare and assign'
	// operator.
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
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
