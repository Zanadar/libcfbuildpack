/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runner

import (
	"os"
	"os/exec"
)

// CommandRunner is an empty struct to hang the Run method on.
type CommandRunner struct {
	Dir string
	Env []string
}

// Run makes CommandRunner satisfy the Runner interface.  This implementation delegates to exec.Command.
func (r CommandRunner) Run(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = r.Dir
	cmd.Env = r.Env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunWithOutput makes CommandRunner satisfy the Runner interface.  This implementation delegates to exec.Command.
func (r CommandRunner) RunWithOutput(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = r.Dir
	cmd.Env = r.Env
	return cmd.CombinedOutput()
}
