// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
    "os/signal"
	"syscall"
	"fmt"
	"github.com/codefresh-io/venona/venonactl/cmd"
)

const (
	waitForSignalEnv       = "WAIT_FOR_DEBUGGER"
	debuggerPort		   = "4321"
)

func main() {
	// Waiting for debugger attach in case if waitForSignalEnv!=""
	// For debuging venonactl spawned by `codefresh runner ...`
	if os.Getenv(waitForSignalEnv) != "" {
		sigs := make(chan os.Signal, 1)
		goOn := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)

		go func() {
			sig := <-sigs
			if sig == syscall.SIGUSR1 {
				goOn <- true
			} else if (sig == syscall.SIGTERM || sig == syscall.SIGINT ){
                fmt.Printf("Exiting ...")
				os.Exit(0)
			}
		}()		
			
		fmt.Printf("%s env is set, waiting SIGUSR1.\nYou can run remote debug in vscode and attach dlv debugger:\n\n", waitForSignalEnv)
	
		pid := os.Getpid()
		fmt.Printf("dlv attach --continue --accept-multiclient --headless --listen=:%s %d\n", debuggerPort, pid)
		fmt.Printf("kill -SIGUSR1 %d\n", pid)
		
		<-goOn
		fmt.Printf("Continue ...")
	}
	cmd.Execute()
}
