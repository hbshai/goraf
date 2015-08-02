package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	serverAddr     = flag.String("addr", "127.0.0.1:8000", "server address (host:port)")
	backupDirPath  = flag.String("backup", "./backup", "path to backup directory")
	filePath       = flag.String("file", "./programs.json", "path to programs.json")
	sessionTimeout = flag.String("timeout", "5m", "session timeout, i.e 3s, 5m10s, etc..")
)

var (
	accessLock         = sync.Mutex{}
	lastAccess         = time.Now()
	lastSessionAddress = net.IPv4(0, 0, 0, 0)
	protectionTime     = time.Second * 20
)

var (
	ErrAccessConflict       = errors.New("Conflicting access")
	ErrDuplicateProgramKeys = errors.New("Duplicate program keys")
)

var (
	infoLog *log.Logger
	errorLog *log.Logger
)

/**
 * Allow http handlers to return errors directly, instead of:
 * 	log(err)
 * 	response.write(error-message)
 * 	return
 * we just do:
 *	return &appError{err, msg, httpCode}
 */
type appError struct {
	Error   error
	Message string
	Code    int
}

func logRequest(r *http.Request, code int) {
	infoLog.Printf("%s %s %d (%s)\n", r.Method, r.URL.Path, code, r.RemoteAddr)
}

// The actual handler to use above struct
type appHandler func(http.ResponseWriter, *http.Request) *appError

// Http handlers must have a ServeHTTP method, so call the actual handler...
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ...here!
	if err := fn(w, r); err != nil {
		logRequest(r, err.Code)
		errorLog.Printf("%v: %s\n", err.Error, err.Message)
		http.Error(w, err.Message, err.Code)
		return
	}
	logRequest(r, 200)
}

// Convert POST data to Program. Convert Program to JSON.
type Program struct {
	Name        string `json:"name"`
	RSS         string `json:"rss"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

/**
 * Check if a request is allowed or blocked.
 * r.RemoteAddr will probably return different addresses for you when
 * developing, due to localhost requests. But it works as intended when
 * a remote machine requests access.
 */
func accessProtected(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return true
	}
	
	// We are the active user
	if ip := net.ParseIP(host); ip.Equal(lastSessionAddress) {
		return false
	}

	return time.Since(lastAccess) < protectionTime
}
func giveAccess(r *http.Request) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	
	lastSessionAddress = net.ParseIP(host)
	lastAccess = time.Now()
}

// https://gist.github.com/elazarl/5507969
func cp(d *os.File, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

/**
 * Will put copies in the backup directory.
 * File names are programs.json-randomNumber
 * Match log statements with file creation dates to find specific backups.
 */
func makeBackup() *appError {
	f, err := ioutil.TempFile(*backupDirPath, "programs.json-")
	if err != nil {
		return &appError{err, "Couldn't create backup file", 0}
	}

	if err = cp(f, *filePath); err != nil {
		return &appError{err, "Failed to copy source to backup", 0}
	}

	return nil
}

// Handles GET and POST requests to /programs. Enforces session access.
func handlePrograms(w http.ResponseWriter, r *http.Request) *appError {
	
	// Use lock to prevent race conditions from multiple goroutines (1/http request)
	accessLock.Lock()
	if accessProtected(r) {
		accessLock.Unlock()

		return &appError{ErrAccessConflict, "there already exists an active session", http.StatusConflict}
	}
	giveAccess(r)
	accessLock.Unlock()

	if r.Method == "GET" {
		data, err := ioutil.ReadFile(*filePath)
		if err != nil {
			return &appError{
				err,
				"Error: Coudln't find/read the programs.json file :(",
				http.StatusInternalServerError}
		}

		w.Write(data)
		return nil
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			return &appError{err,
				"Error: Couldn't parse form data, contact the IT monkey",
				http.StatusBadRequest}
		}

		// Everything is posted as arrays...
		keys := r.Form["programs[][key]"]
		names := r.Form["programs[][name]"]
		rss := r.Form["programs[][rss]"]
		images := r.Form["programs[][image]"]
		categories := r.Form["programs[][category]"]
		descriptions := r.Form["programs[][description]"]

		// So generate the JSON-like object
		programs := make(map[string]Program)
		for i := 0; i < len(keys); i++ {
			if _, ok := programs[keys[i]]; ok {
				return &appError{ErrDuplicateProgramKeys,
					fmt.Sprintf("Error: program key '%s' already exists (check for duplicates)", keys[i]),
					http.StatusBadRequest}
			}

			programs[keys[i]] = Program{
				Name:        names[i],
				RSS:         rss[i],
				Image:       images[i],
				Category:    categories[i],
				Description: descriptions[i],
			}
		}

		// And then output nicely formatted JSON
		data, err := json.MarshalIndent(programs, "", "   ")
		if err != nil {
			return &appError{err,
				"Error: Couldn't convert data to json, contact IT turtle",
				http.StatusInternalServerError}
		}

		// Try to make a backup for safety purposes, but don't enforce it
		backupErr := makeBackup()
		if backupErr != nil {
			errorLog.Printf("%s: %v\n", backupErr.Message, backupErr.Error)
		}

		// We don't need executable rights, just read-write
		if err = ioutil.WriteFile(*filePath, data, 0644); err != nil {
			return &appError{err,
				"Error: Couldn't write to programs.json, contact IT shibe",
				http.StatusInternalServerError}
		}

		infoLog.Printf("Wrote %d bytes\n", len(data))
		if backupErr == nil {
			fmt.Fprintf(w, "Success: %d programs saved (~%d kB)", len(programs), len(data)/1000)
		} else {
			fmt.Fprintf(w, "Warning: %d programs saved (~%d kB), but failed to backup.", len(programs), len(data)/1000)
		}
	}

	return nil
}

// Probe current session status
func handleAccess(w http.ResponseWriter, r *http.Request) *appError {
	accessLock.Lock()
	if accessProtected(r) {
		accessLock.Unlock()

		// The client expects the number of seconds left until session timeout.
		dur := fmt.Sprintf("%d", int((protectionTime - time.Since(lastAccess)).Seconds()))
		return &appError{ErrAccessConflict, dur, http.StatusConflict}
	}
	accessLock.Unlock()
	fmt.Fprintf(w, "You can get access!")
	return nil
}

func main() {
	flag.Parse()

	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Check if program.json file exists, to prevent read errors later on 
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		errorLog.Fatalln("Couldn't find json file in " + *filePath)
	}
	infoLog.Println("Program file path is " + *filePath)

	// Try to make sure that we habe a backup dir to store copies in
	if _, err := os.Stat(*backupDirPath); os.IsNotExist(err) {
		infoLog.Println("Creating backup directory in " + *backupDirPath)
		if err := os.Mkdir(*backupDirPath, 0744); err != nil {
			errorLog.Printf("Couldn't create backup dir. (%v)\n", err)
		}
	}
	infoLog.Println("Backup directory is " + *backupDirPath)

	// Do assignment so that we don't shadow the global...
	var err error
	protectionTime, err = time.ParseDuration(*sessionTimeout)
	if err != nil {
		errorLog.Printf("Failed to parse session time, falling back to: %v\n", err)
		protectionTime = 5 * time.Minute
	}

	// Better trick than checking for "no one"
	lastAccess = lastAccess.Add(-protectionTime)
	infoLog.Printf("Session timeout is %v\n", protectionTime)

	// Host files from public dir, but mount them on root (/) path instead
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.Handle("/programs", appHandler(handlePrograms))
	http.Handle("/access", appHandler(handleAccess))

	infoLog.Printf("Running http server on address %s\n", *serverAddr)
	http.ListenAndServe(*serverAddr, nil)
}
