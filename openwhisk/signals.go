package openwhisk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"syscall"
)

func (ap *ActionProxy) signalHandler(w http.ResponseWriter, r *http.Request, signal syscall.Signal) {

	// parse the request
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	Debug("done reading %d bytes", len(body))

	// check if you have an action
	if ap.theExecutor == nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("no action defined yet"))
		return
	}
	// check if the process exited
	if ap.theExecutor.Exited() {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("command exited"))
		return
	}

	// remove newlines
	body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

	// execute the action
	response, err := ap.theExecutor.Singal(signal)

	// check for early termination
	if err != nil {
		Debug("WARNING! Command exited")
		ap.theExecutor = nil
		sendError(w, http.StatusBadRequest, fmt.Sprintf("command exited"))
		return
	}
	DebugLimit("received:", response, 120)

	// check if the answer is an object map
	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(response, &objmap)
	if err != nil {
		sendError(w, http.StatusBadGateway, "The action did not return a dictionary.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
	numBytesWritten, err := w.Write(response)

	// flush output
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// diagnostic when you have writing problems
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
		return
	}
	if numBytesWritten != len(response) {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
		return
	}
}
