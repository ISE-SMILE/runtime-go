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
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ackMsg struct {
	Ok        bool `json:"ok"`
	PauseOk   bool `json:"pause,omitempty"`
	FinishOk  bool `json:"finish,omitempty"`
	HintOk    bool `json:"hint,omitempty"`
	FreshenOk bool `json:"freshen,omitempty"`
}

func init() {
	zerolog.TimeFieldFormat = ""
}

func main() {

	// assign the main function
	type Action func(event map[string]interface{}) map[string]interface{}
	var action Action
	action = Main

	// input
	out := os.NewFile(3, "pipe")
	defer out.Close()
	reader := bufio.NewReader(os.Stdin)

	capture := make(chan os.Signal, 2)
	signal.Notify(capture, syscall.SIGINT, syscall.SIGABRT, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for {
			sig := <-capture
			fmt.Printf("{\"signal\":\"%+v\"}%s", sig, '\n')
			fmt.Fprintf(out, "{\"signal\":\"%+v\"}\n", sig)

			if sig == syscall.SIGTRAP {

				return
			}
		}
	}()

	msg := ackMsg{
		Ok:        true,
		PauseOk:   true,
		FinishOk:  true,
		HintOk:    true,
		FreshenOk: true,
	}
	buf, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintf(out, `{ \"ok\": false , }%s`, "\n")
	} else {
		fmt.Fprintln(out, string(buf))
	}

	for {
		// read one line
		inbuf, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		// parse one line
		var input map[string]interface{}
		err = json.Unmarshal(inbuf, &input)
		if err != nil {
			fmt.Fprintf(out, "{ error: %q}\n", err.Error())
			continue
		}

		// set environment variables
		err = json.Unmarshal(inbuf, &input)
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
			fmt.Fprintf(out, "{ error: %q}\n", err.Error())
			continue
		}
		output = bytes.Replace(output, []byte("\n"), []byte(""), -1)
		fmt.Fprintf(out, "%s\n", output)
	}
}

func Main(obj map[string]interface{}) map[string]interface{} {
	name, ok := obj["name"].(string)
	if !ok {
		name = "world"
	}
	log.Debug().Str("name", name).Msg("Hello")
	msg := make(map[string]interface{})
	msg["module-main"] = "Hello, " + name + "!"
	return msg
}
