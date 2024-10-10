// Copyright 2024 Qian Yao
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metacmd

import "time"

// ExecType represents the type of execution requested.
type ExecType int

const (
	// ExecNone indicates no execution.
	ExecNone ExecType = iota
	// ExecOnly indicates plain execution only (\g).
	ExecOnly
	// ExecPipe indicates execution and piping results (\g |file)
	ExecPipe
	// ExecSet indicates execution and setting the resulting columns as
	// variables (\gset).
	ExecSet
	// ExecExec indicates execution and executing the resulting rows (\gexec).
	ExecExec
	// ExecCrosstab indicates execution using crosstabview (\crosstabview).
	ExecCrosstab
	// ExecChart indicates execution using chart (\chart).
	ExecChart
	// ExecWatch indicates repeated execution with a fixed time interval.
	ExecWatch
)

// Option contains parsed result options of a metacmd.
type Option struct {
	// Quit instructs the handling code to quit.
	Quit bool
	// Exec informs the handling code of the type of execution.
	Exec ExecType
	// Params are accompanying string parameters for execution.
	Params map[string]string
	// Crosstab are the crosstab column parameters.
	Crosstab []string
	// Watch is the watch duration interval.
	Watch time.Duration
}
