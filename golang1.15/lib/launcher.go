/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// OwExecutionEnv is the execution environment set at compile time
var OwExecutionEnv = ""

type ackMsg struct {
	Ok        bool  `json:"ok"`
	PauseOk   bool `json:"pause,omitempty"`
	FinishOk  bool `json:"finish,omitempty"`
	HintOk    bool `json:"hint,omitempty"`
	FreshenOk bool `json:"freshen,omitempty"`
}


type signals interface {
	Pause(out *os.File)
	Stop(out *os.File)
	Hint(in map[string]string, out *os.File)
	Freshen(in map[string]string, out *os.File)
}

type BaseSignals struct {}

func (b BaseSignals) Pause(out *os.File) {}

func (b BaseSignals) Stop(out *os.File) {}

func (b BaseSignals) Hint(in map[string]string, out *os.File) {}

func (b BaseSignals) Freshen(in map[string]string, out *os.File) {}

var Interruppts signals

//pause,stop,hint,freshen
var InterruptSupport ackMsg = ackMsg{
	Ok:        true,
	PauseOk:   false,
	FinishOk:  false,
	HintOk:    false,
	FreshenOk: false,
}


func main() {
	// check if the execution environment is correct
	if OwExecutionEnv != "" && OwExecutionEnv != os.Getenv("__OW_EXECUTION_ENV") {
		fmt.Println("Execution Environment Mismatch")
		fmt.Println("Expected: ", OwExecutionEnv)
		fmt.Println("Actual: ", os.Getenv("__OW_EXECUTION_ENV"))
		os.Exit(1)
	}

	// debugging
	var debug = os.Getenv("OW_DEBUG") != ""
	if debug {
		f, err := os.OpenFile("/tmp/action.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(f)
		}
		log.Printf("Environment: %v", os.Environ())
	}

	// assign the main function
	type Action func(event map[string]interface{}) map[string]interface{}
	var action Action
	action = Main

	// input
	out := os.NewFile(3, "pipe")
	defer out.Close()
	reader := bufio.NewReader(os.Stdin)

	if Interruppts != nil{
		signals := make([]os.Signal,0)
		if InterruptSupport.PauseOk {
			signals = append(signals, syscall.SIGINT)
		}

		if InterruptSupport.FinishOk {
			signals = append(signals, syscall.SIGABRT)
		}

		if InterruptSupport.HintOk {
			signals = append(signals, syscall.SIGUSR1)
		}

		if InterruptSupport.FreshenOk {
			signals = append(signals, syscall.SIGUSR2)
		}

		capture := make(chan os.Signal, 2)
		signal.Notify(capture, signals...)


		go func() {
			for {
				sig := <-capture
				switch sig {
				case syscall.SIGINT:
					Interruppts.Pause(out)
				case syscall.SIGABRT:
					Interruppts.Stop(out)
					return
				case syscall.SIGUSR1:
					Interruppts.Hint(nil,out)
				case syscall.SIGUSR2:
					Interruppts.Freshen(nil,out)
				}
			}
		}()
	}

	// acknowledgement of started action
	buf,err := json.Marshal(InterruptSupport)
	if err != nil{
		fmt.Fprintf(out, `{ \"ok\": false , }%s`, "\n")
		return
	} else {
		fmt.Fprintln(out,string(buf))
	}
	if debug {
		log.Println("action started")
	}

	// read-eval-print loop
	for {
		// read one line
		inbuf, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		if debug {
			log.Printf(">>>'%s'>>>", inbuf)
		}
		// parse one line
		var input map[string]interface{}
		err = json.Unmarshal(inbuf, &input)
		if err != nil {
			log.Println(err.Error())
			fmt.Fprintf(out, "{ error: %q}\n", err.Error())
			continue
		}
		if debug {
			log.Printf("%v\n", input)
		}
		// set environment variables
		for k, v := range input {
			if k == "value" {
				continue
			}
			if s, ok := v.(string); ok {
				os.Setenv("__OW_"+strings.ToUpper(k), s)
			}
		}
		// get payload if not empty
		var payload map[string]interface{}
		if value, ok := input["value"].(map[string]interface{}); ok {
			payload = value
		}
		// process the request
		result := action(payload)
		// encode the answer
		output, err := json.Marshal(&result)
		if err != nil {
			log.Println(err.Error())
			fmt.Fprintf(out, "{ error: %q}\n", err.Error())
			continue
		}
		output = bytes.Replace(output, []byte("\n"), []byte(""), -1)
		if debug {
			log.Printf("<<<'%s'<<<", output)
		}
		fmt.Fprintf(out, "%s\n", output)
	}
}
