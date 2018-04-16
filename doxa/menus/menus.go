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

package menus

import (
	"sort"
	"strconv"
	"sync"

	"github.com/beevik/etree"
	"github.com/labneco/doxa/doxa/actions"
)

// Registry is the menu Collection of the application
var (
	Registry     *Collection
	bootstrapMap map[string]*Menu
)

// A Collection is a hierarchical and sortable Collection of menus
type Collection struct {
	sync.RWMutex
	Menus    []*Menu
	menusMap map[string]*Menu
}

func (mc *Collection) Len() int {
	return len(mc.Menus)
}

func (mc *Collection) Swap(i, j int) {
	mc.Menus[i], mc.Menus[j] = mc.Menus[j], mc.Menus[i]
}

func (mc *Collection) Less(i, j int) bool {
	return mc.Menus[i].Sequence < mc.Menus[j].Sequence
}

// Add adds a menu to the menu Collection
func (mc *Collection) Add(m *Menu) {
	if m.Action != nil {
		m.HasAction = true
	}
	var targetCollection *Collection
	if m.Parent != nil {
		if m.Parent.Children == nil {
			m.Parent.Children = NewCollection()
		}
		targetCollection = m.Parent.Children
		m.Parent.HasChildren = true
	} else {
		targetCollection = mc
	}
	m.ParentCollection = targetCollection
	targetCollection.Menus = append(targetCollection.Menus, m)
	sort.Sort(targetCollection)

	// We add the menu to the Registry which is the top collection
	mc.Lock()
	defer mc.Unlock()
	Registry.menusMap[m.ID] = m
}

// GetByID returns the Menu with the given id
func (mc *Collection) GetByID(id string) *Menu {
	return mc.menusMap[id]
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		menusMap: make(map[string]*Menu),
	}
	return &res
}

// A Menu is the representation of a single menu item
type Menu struct {
	ID               string
	Name             string
	ParentID         string
	Parent           *Menu
	ParentCollection *Collection
	Children         *Collection
	Sequence         uint8
	ActionID         string
	Action           *actions.Action
	HasChildren      bool
	HasAction        bool
	names            map[string]string
}

// TranslatedName returns the translated name of this menu
// in the given language
func (m Menu) TranslatedName(lang string) string {
	res, ok := m.names[lang]
	if !ok {
		res = m.Name
	}
	return res
}

// LoadFromEtree reads the menu given etree.Element, creates or updates the menu
// and adds it to the menu registry if it not already.
func LoadFromEtree(element *etree.Element) {
	AddMenuToMapFromEtree(element, bootstrapMap)
}

// AddMenuToMapFromEtree reads the menu from the given element
// and adds it to the given map.
func AddMenuToMapFromEtree(element *etree.Element, mMap map[string]*Menu) map[string]*Menu {
	seq, _ := strconv.Atoi(element.SelectAttrValue("sequence", "10"))
	menu := Menu{
		ID:       element.SelectAttrValue("id", "NO_ID"),
		ActionID: element.SelectAttrValue("action", ""),
		Name:     element.SelectAttrValue("name", ""),
		ParentID: element.SelectAttrValue("parent", ""),
		Sequence: uint8(seq),
	}
	mMap[menu.ID] = &menu
	return mMap
}
