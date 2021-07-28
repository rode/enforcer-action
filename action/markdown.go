// Copyright 2021 The Rode Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package action

import (
	"fmt"
	"strings"
)

type markdownPrinter struct {
	builder strings.Builder
}

func (md *markdownPrinter) h1(title string, values ...interface{}) *markdownPrinter {
	md.header(1, title, values...)

	return md
}

func (md *markdownPrinter) h2(title string, values ...interface{}) *markdownPrinter {
	md.header(2, title, values...)

	return md
}

func (md *markdownPrinter) h3(title string, values ...interface{}) *markdownPrinter {
	md.header(3, title, values...)

	return md
}

func (md *markdownPrinter) comment(message string) *markdownPrinter {
	fmt.Fprint(&md.builder, "<!---")
	fmt.Fprint(&md.builder, message)
	fmt.Fprintln(&md.builder, "--->")

	return md
}

func (md *markdownPrinter) header(depth int, title string, values ...interface{}) {
	for i := 0; i < depth; i++ {
		fmt.Fprint(&md.builder, "#")
	}

	fmt.Fprint(&md.builder, " ")
	fmt.Fprintf(&md.builder, title, values...)
	md.newline()
	md.newline()
}

func (md *markdownPrinter) table(headers []string, rows [][]string) *markdownPrinter {
	for _, h := range headers {
		md.tableRowEntry(h)
	}

	md.newline()

	for range headers {
		md.tableRowEntry("--")
	}

	md.newline()

	for _, row := range rows {
		for _, entry := range row {
			md.tableRowEntry(entry)
		}
		md.newline()
	}

	md.newline()

	return md
}

func (md *markdownPrinter) tableRowEntry(entry string) {
	fmt.Fprint(&md.builder, "| ")
	fmt.Fprint(&md.builder, entry)
	fmt.Fprint(&md.builder, " |")
}

func (md *markdownPrinter) codeBlock() *markdownPrinter {
	fmt.Fprintln(&md.builder, "```")

	return md
}

func (md *markdownPrinter) quote(message string) *markdownPrinter {
	fmt.Fprint(&md.builder, "> ")
	fmt.Fprintln(&md.builder, message)

	return md
}

func (md *markdownPrinter) write(line string) *markdownPrinter {
	fmt.Fprintln(&md.builder, line)

	return md
}

func (md *markdownPrinter) newline() *markdownPrinter {
	fmt.Fprintln(&md.builder, "")

	return md
}

func (md *markdownPrinter) list(items []string) *markdownPrinter {
	for _, item := range items {
		fmt.Fprint(&md.builder, "- ")
		fmt.Fprintln(&md.builder, item)
	}

	return md
}

func (md *markdownPrinter) string() string {
	return md.builder.String()
}

func asCode(line string) string {
	return fmt.Sprintf("`%s`", line)
}
