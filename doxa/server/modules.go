// Copyright 2016 NDP Systèmes. All Rights Reserved.
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

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/beevik/etree"
	"github.com/labneco/doxa/doxa/actions"
	"github.com/labneco/doxa/doxa/i18n"
	"github.com/labneco/doxa/doxa/menus"
	"github.com/labneco/doxa/doxa/models"
	"github.com/labneco/doxa/doxa/tools/generate"
	"github.com/labneco/doxa/doxa/views"
)

// A Module is a go package that implements business features.
// This struct is used to register modules.
type Module struct {
	Name     string
	PreInit  func()
	PostInit func()
}

// A ModulesList is a list of Module objects
type ModulesList []*Module

// Names returns a list of all module names in this ModuleList.
func (ml *ModulesList) Names() []string {
	res := make([]string, len(*ml))
	for i, module := range *ml {
		res[i] = module.Name
	}
	return res
}

// Modules is the list of activated modules in the application
var Modules ModulesList

// RegisterModule registers the given module in the server
// This function should be called in the init() function of
// all Doxa Addons.
func RegisterModule(mod *Module) {
	Modules = append(Modules, mod)
}

// LoadInternalResources loads all data in the 'resources' directory, that are
// - views,
// - actions,
// - menu items
// Internal resources are defined in XML files.
func LoadInternalResources() {
	loadData("resources", "xml", loadXMLResourceFile)
}

// LoadDataRecords loads all the data records in the 'data' directory into the database.
// Data records are defined in CSV files.
func LoadDataRecords() {
	loadData("data", "csv", models.LoadCSVDataFile)
}

// LoadDemoRecords loads all the data records in the 'demo' directory into the database.
// Demo records are defined in CSV files.
func LoadDemoRecords() {
	loadData("demo", "csv", models.LoadCSVDataFile)
}

// LoadTranslations loads all translation data from the PO files in the 'i18n' directory
// into the translations registry.
func LoadTranslations(langs []string) {
	for _, mod := range Modules {
		dataDir := filepath.Join(generate.DoxaDir, "doxa", "server", "i18n", mod.Name)
		if _, err := os.Stat(dataDir); err != nil {
			// No resources dir in this module
			return
		}
		LoadModuleTranslations(dataDir, langs)
	}
}

// LoadModuleTranslations loads the PO files in the given directory for the given languages
func LoadModuleTranslations(i18nDir string, langs []string) {
	var poFiles []string
	for _, lang := range langs {
		dataFiles, err := filepath.Glob(fmt.Sprintf("%s/%s.po", i18nDir, lang))
		if err != nil {
			log.Panic("Unable to scan directory for data files", "dir", i18nDir, "type", "po", "error", err)
		}
		poFiles = append(poFiles, dataFiles...)
	}
	dataFilesSorted := sort.StringSlice(poFiles)
	dataFilesSorted.Sort()
	for _, dataFile := range dataFilesSorted {
		i18n.LoadPOFile(dataFile)
	}
}

// loadData loads the files in the given dir with the given extension (without .)
// using the loader function.
func loadData(dir, ext string, loader func(string)) {
	for _, mod := range Modules {
		dataDir := filepath.Join(generate.DoxaDir, "doxa", "server", dir, mod.Name)
		if _, err := os.Stat(dataDir); err != nil {
			// No resources dir in this module
			continue
		}
		dataFiles, err := filepath.Glob(fmt.Sprintf("%s/*.%s", dataDir, ext))
		if err != nil {
			log.Panic("Unable to scan directory for data files", "dir", dataDir, "type", ext, "error", err)
		}
		dataFilesSorted := sort.StringSlice(dataFiles)
		dataFilesSorted.Sort()
		for _, dataFile := range dataFilesSorted {
			loader(dataFile)
		}
	}
}

// loadXMLResourceFile loads the data from an XML data file into memory.
func loadXMLResourceFile(fileName string) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fileName); err != nil {
		log.Panic("Error loading XML data file", "file", fileName, "error", err)
	}
	for _, dataTag := range doc.FindElements("doxa/data") {
		for _, object := range dataTag.ChildElements() {
			switch object.Tag {
			case "view":
				views.LoadFromEtree(object)
			case "action":
				actions.LoadFromEtree(object)
			case "menuitem":
				menus.LoadFromEtree(object)
			default:
				log.Panic("Unknown XML tag", "filename", fileName, "tag", object.Tag)
			}
		}
	}
}
