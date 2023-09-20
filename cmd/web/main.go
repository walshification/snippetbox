package main

import (
	"flag"
	"log"
	"net/http"
)

type config struct {
	addr      string
	staticDir string
}

func main() {
	var cfg config

	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the
	// flag will be stored in the addr variable at runtime.
	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	// Do the same for our static assets path.
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")

	// Importantly, we use the flag.Parse() function to parse the command-line flag.
	// This reads in the command-line flag value and assigns it to the addr
	// variable. You need to call this *before* you use the addr variable
	// otherwise it will always contain the default value of ":4000". If any errors are
	// encountered during parsing the application will be terminated.
	flag.Parse()

	// Use the http.NewServeMux() function to initialize a new servemux.
	mux := http.NewServeMux()

	// Create a file server which serves files out of the "./ui/static" directory.
	// Note that the path given to the http.Dir function is relative to the project
	// directory root.
	fileServer := http.FileServer(http.Dir(cfg.staticDir))

	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/". For matching paths, we strip the
	// "/static" perfix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Register the home function as the handler for the "/" URL pattern.
	mux.HandleFunc("/", home)

	// Register the two new handler functions and corresponding URL patters with
	// the servemux, in exactly the same way that we did before.
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	// The value returned from the flag.String() function is a pointer to the flag
	// value, not the value itself. So we need to dereference the pointer (i.e.
	// prefix it with the * symbol) before using it. Note that we're using the
	// log.Printf() function to interpolate the address with the log message.
	log.Printf("Starting server on %s", cfg.addr)

	// Use the http.ListenAndServe() function to start a new web server. We pass in
	// two parameters: the TCP network address to listen on (in this case ":4000")
	// and the servemux we just created. If http.ListenAndServe() returns an error
	// we use the log.Fatal() function to log the error message and exit. Note
	// that any error returned by http.ListenAndServe() is always non-nil.
	err := http.ListenAndServe(cfg.addr, mux)
	log.Fatal(err)
}
