// Copyright 2026 Google LLC
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

// The validate_catalog command validates Hugo shortcodes, Markdown links and
// anchors, and multi-language code snippets in the Cloud Samples Catalog
// repository.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		if !errors.Is(err, cmd.ErrValidationFailed) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}
