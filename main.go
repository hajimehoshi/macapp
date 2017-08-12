// Copyright 2017 Hajime Hoshi
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
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const infoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>CFBundleGetInfoString</key>
    <string>{{.AppName}}</string>
    <key>CFBundleExecutable</key>
    <string>{{.AppName}}</string>
    <key>CFBundleIdentifier</key>
    <string>com.example.www</string>
    <key>CFBundleName</key>
    <string>{{.AppName}}</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundleShortVersionString</key>
    <string>0.01</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>IFMajorVersion</key>
    <integer>0</integer>
    <key>IFMinorVersion</key>
    <integer>1</integer>
    <key>NSHighResolutionCapable</key>
    <true />
  </dict>
</plist>
`

var (
	output = flag.String("o", "", "application name")
)

func escapeXML(str string) string {
	buf := &bytes.Buffer{}
	xml.EscapeText(buf, []uint8(str))
	return buf.String()
}

func run(dir string, appName string, bin string) error {
	const (
		perm      = 0755
		ext       = ".app"
		contents  = "Contents"
		macOS     = "MacOS"
		resources = "Resources"
	)
	appPath := filepath.Join(dir, appName+ext)

	// Delete .app if it already exists.
	if err := os.RemoveAll(appPath); err != nil {
		return err
	}

	// Create directories.
	if err := os.MkdirAll(filepath.Join(appPath, contents, macOS), perm); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(appPath, contents, resources), perm); err != nil {
		return err
	}

	// Write the Info.plist file.
	w := []uint8(strings.Replace(infoPlist, "{{.AppName}}", escapeXML(appName), -1))
	if err := ioutil.WriteFile(filepath.Join(appPath, contents, "Info.plist"), w, 0644); err != nil {
		return err
	}

	// Copy the binary file.
	in, err := os.Open(bin)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(filepath.Join(appPath, contents, macOS, appName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if *output == "" {
		fmt.Fprintf(os.Stderr, "application path/name is not specified\n")
		flag.Usage()
		os.Exit(1)
	}

	bin := flag.Arg(0)
	if bin == "" {
		fmt.Fprintf(os.Stderr, "binary is not specified\n")
		os.Exit(1)
	}

	if filepath.Ext(*output) != ".app" {
		fmt.Fprintf(os.Stderr, "application name must end with .app\n")
		os.Exit(1)
	}
	appName := filepath.Base(*output)
	appName = appName[:len(appName)-4]
	dir := filepath.Dir(*output)

	if err := run(dir, appName, bin); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
