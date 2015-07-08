package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	serverPort     = flag.String("addr", "127.0.0.1:8000", "server address (host:port)")
	backupDirPath  = flag.String("backup", "./backup", "path to backup directory")
	filePath       = flag.String("file", "./programs.json", "path to programs.json")
	sessionTimeout = flag.String("timeout", "5m", "session timeout, i.e 3s, 5m10s, etc..")

	lastAccess         = time.Now()
	lastSessionAddress = "no one"
	protectionTime     = time.Second * 20
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

// The actual handler to use above struct
type appHandler func(http.ResponseWriter, *http.Request) *appError

// Http handlers must have a ServeHTTP method, so call the actual handler...
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ...here!
	if err := fn(w, r); err != nil {
		log.Printf("%s: %v\n", err.Message, err.Error)
		http.Error(w, err.Message, err.Code)
	}
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
	// We are the active user
	if r.RemoteAddr == lastSessionAddress {
		return false
	}

	return time.Since(lastAccess) < protectionTime
}
func giveAccess(r *http.Request) {
	lastSessionAddress = r.RemoteAddr
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
	if accessProtected(r) {
		// The client expects the number of seconds left until session timeout.
		dur := fmt.Sprintf("%d", int((protectionTime - time.Since(lastAccess)).Seconds()))
		return &appError{errors.New("Conflicting access"), dur, 400}
	}
	giveAccess(r)

	if r.Method == "GET" {
		data, err := ioutil.ReadFile(*filePath)
		if err != nil {
			return &appError{err, "Coudln't find/read the programs.json file :(", 500}
		}

		w.Write(data)
		return nil
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			return &appError{err, "Error: Couldn't parse form data, check log and contact the IT monkey", 400}
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
			return &appError{err, "Error: Couldn't convert data to json, check log and contact IT turtle", 500}
		}
		
		// Try to make a backup for safety purposes, but don't enforce it
		backupErr := makeBackup()
		if backupErr != nil {
			log.Printf("%s: %v\n", backupErr.Message, backupErr.Error)
		}
		
		// We don't need executable rights, just read-write
		if err = ioutil.WriteFile(*filePath, data, 0644); err != nil {
			return &appError{err, "Couldn't write to programs.json, perhaps some premission error?", 500}
		}
		
		log.Printf("Wrote %d bytes\n", len(data))
		if backupErr == nil {
			fmt.Fprintf(w, "Success: %d programs saved (~%d kB)", len(programs), len(data)/1000)
		} else {
			fmt.Fprintf(w, "Warning: %d programs saved (~%d kB), but failed to backup.", len(programs), len(data)/1000)
		}
	}

	return nil
}

func main() {
	flag.Parse()
	
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalln("Coudln't find json file in " + *filePath)
	}
	log.Println("Program file path is " + *filePath)

	if _, err := os.Stat(*backupDirPath); os.IsNotExist(err) {
		log.Println("Creating backup directory in " + *backupDirPath)
		if err := os.Mkdir(*backupDirPath, 0744); err != nil {
			log.Printf("Couldn't create backup dir. Error: %v", err)
		}
	}
	log.Println("Backup directory is " + *backupDirPath)

	// Do assignment so that we don't shadow the global...
	var err error
	protectionTime, err = time.ParseDuration(*sessionTimeout)
	if err != nil {
		log.Printf("Failed to parse session time, falling back...%v\n", err)
		protectionTime = 5 * time.Minute
	}

	// Better trick than checking for "no one"
	lastAccess = lastAccess - protectionTime
	log.Printf("Session timeout is %v\n", protectionTime)

	// Host files from public dir, but mount them on root (/) path instead
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", http.StripPrefix("/", fs))

	http.Handle("/programs", appHandler(handlePrograms))

	log.Printf("Running http server on address %s\n", *serverPort)
	http.ListenAndServe(*serverPort, nil)
}