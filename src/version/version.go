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

package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	Version   = "0.0.1"
	Commit    string
	BuildDate string
)

// GetVersionDetail returns a string with version, commit hash, build time, and OS/Arch details.
func GetVersionDetail() string {
	versionDetail := fmt.Sprintf(`
version:   %s
commit:    %s
buildtime: %s
os/arch:   %s/%s`,
		Version,
		Commit,
		BuildDate,
		runtime.GOOS,
		runtime.GOARCH)

	return strings.TrimSpace(versionDetail)
}
