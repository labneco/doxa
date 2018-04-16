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

package security

import (
	"fmt"
	"sync"
)

const (
	// SuperUserID is the uid of the administrator
	SuperUserID int64 = 1
	// GroupAdminID is the string ID of the group with all permissions
	GroupAdminID string = "admin"
	// GroupEveryoneID is the string ID of the group everyone belongs to
	GroupEveryoneID string = "everyone"

	// NativeGroup means that this user has been explicitly given membership in this group
	NativeGroup InheritanceInfo = iota
	// InheritedGroup means that this user is a member of this group through inheritance
	InheritedGroup
)

// Registry of all security groups of the application
var Registry *GroupCollection

var (
	// GroupAdmin which has all permissions
	GroupAdmin *Group
	// GroupEveryone is a group that all users automatically belong to.
	GroupEveryone *Group
)

// InheritanceInfo enables us to know if a user is part of a group
// natively or by inheritance.
type InheritanceInfo int8

// A Group defines a role which can be granted or denied permissions.
// - Groups can inherit from other groups and get access to these groups
// permissions.
// - A user can belong to one or several groups, and thus inherit from the
// permissions of the groups.
type Group struct {
	ID       string
	Name     string
	Inherits []*Group
}

// String function for group
func (g Group) String() string {
	return fmt.Sprintf("Group(%s)", g.ID)
}

// A GroupCollection keeps a list of groups
type GroupCollection struct {
	sync.RWMutex
	groups      map[string]*Group
	memberships map[int64]map[*Group]InheritanceInfo
}

// NewGroup creates a new Group with the given id, name and inherited groups
// and registers it in this GroupCollection. It returns a pointer to the newly
// created group.
func (gc *GroupCollection) NewGroup(ID, name string, inherits ...*Group) *Group {
	grp := &Group{
		ID:       ID,
		Name:     name,
		Inherits: inherits,
	}
	gc.RegisterGroup(grp)
	return grp
}

// RegisterGroup adds the given group to this GroupCollection
// If group with the same ID exists, this methods panics.
func (gc *GroupCollection) RegisterGroup(group *Group) {
	gc.Lock()
	defer gc.Unlock()
	if _, exists := gc.groups[group.ID]; exists {
		log.Panic("Trying register a new group with an existing ID", "ID", group.ID)
	}
	gc.groups[group.ID] = group
}

// inheritedBy recursively populates the result slice for the
// with the group's parents
func (gc *GroupCollection) inheritedBy(group *Group, result *[]*Group) {
	for _, parent := range group.Inherits {
		*result = append(*result, parent)
		gc.inheritedBy(parent, result)
	}
}

// UnregisterGroup removes the group with the given ID from this GroupCollection
func (gc *GroupCollection) UnregisterGroup(group *Group) {
	// remove links from inheriting groups
	for id, grp := range gc.groups {
		for i, iGrp := range grp.Inherits {
			if iGrp.ID == group.ID {
				// memory safe delete
				copy(gc.groups[id].Inherits[i:], gc.groups[id].Inherits[i+1:])
				length := len(gc.groups[id].Inherits)
				gc.groups[id].Inherits[length-1] = nil
				gc.groups[id].Inherits = gc.groups[id].Inherits[:length-1]
			}
		}
	}
	// remove memberships
	for uid := range gc.memberships {
		gc.RemoveMembership(uid, group)
	}
	// Remove the group itself
	gc.Lock()
	defer gc.Unlock()
	delete(gc.groups, group.ID)
}

// GetGroup returns the group with the given groupID or nil if not found
func (gc *GroupCollection) GetGroup(groupID string) *Group {
	return gc.groups[groupID]
}

// AddMembership adds the user defined by its uid to the
// given group and also to all groups that inherit this group.
// inherit is set to true when this method is called on an
// inherited group recursively. You should normally leave it
// unset.
func (gc *GroupCollection) AddMembership(uid int64, group *Group, inherit ...bool) {
	var inheritingGroups []*Group
	gc.inheritedBy(group, &inheritingGroups)
	for _, grp := range inheritingGroups {
		gc.AddMembership(uid, grp, true)
	}
	gc.Lock()
	defer gc.Unlock()
	mode := NativeGroup
	if len(inherit) > 0 && inherit[0] {
		mode = InheritedGroup
	}
	if _, exists := gc.memberships[uid]; !exists {
		gc.memberships[uid] = make(map[*Group]InheritanceInfo)
	}
	gc.memberships[uid][group] = mode
}

// RemoveMembership removes the user with the given uid from the given group
// and all groups that inherit from this group.
func (gc *GroupCollection) RemoveMembership(uid int64, group *Group) {
	if _, exists := gc.memberships[uid][group]; !exists {
		return
	}
	gc.doRemoveMembership(uid, group)
	// Re-Add membership for all existing groups to compute inheritance
	for grp, ii := range gc.memberships[uid] {
		if ii == NativeGroup {
			gc.AddMembership(uid, grp)
		}
	}
}

// doRemoveMembership actually removes the user with the given uid from the
// given Group and all groups that inherit from this Group.
func (gc *GroupCollection) doRemoveMembership(uid int64, group *Group) {
	gc.Lock()
	defer gc.Unlock()
	// Remove our group
	delete(gc.memberships[uid], group)
	// Remove all inherited groups
	for _, grp := range group.Inherits {
		if gc.memberships[uid][grp] == InheritedGroup {
			delete(gc.memberships[uid], grp)
		}
	}
}

// RemoveAllMembershipsForUser removes the given uid from all groups
func (gc *GroupCollection) RemoveAllMembershipsForUser(uid int64) {
	gc.doRemoveAllMembershipsForUser(uid)
	if uid == SuperUserID {
		gc.AddMembership(SuperUserID, GroupAdmin)
	}
}

// doRemoveAllMembershipsForUser actually removes the given uid from all groups
func (gc *GroupCollection) doRemoveAllMembershipsForUser(uid int64) {
	gc.Lock()
	defer gc.Unlock()
	delete(gc.memberships, uid)
}

// HasMembership returns true id the given uid is a member of the given group
func (gc *GroupCollection) HasMembership(uid int64, group *Group) bool {
	if group == GroupEveryone {
		return true
	}
	_, ok := gc.memberships[uid][group]
	return ok
}

// UserGroups returns the slice of groups the user with the given
// uid belongs to, including inherited groups.
func (gc *GroupCollection) UserGroups(uid int64) map[*Group]InheritanceInfo {
	res := make(map[*Group]InheritanceInfo, len(gc.memberships[uid])+1)
	for k, v := range gc.memberships[uid] {
		res[k] = v
	}
	res[GroupEveryone] = NativeGroup
	return res
}

// AllGroups returns a slice with all the groups of the collection
func (gc *GroupCollection) AllGroups() []*Group {
	res := make([]*Group, len(gc.groups))
	i := 0
	for _, group := range gc.groups {
		res[i] = group
		i++
	}
	return res
}

// NewGroupCollection returns a pointer to a new empty GroupCollection
func NewGroupCollection() *GroupCollection {
	gc := GroupCollection{
		groups:      make(map[string]*Group),
		memberships: make(map[int64]map[*Group]InheritanceInfo),
	}
	return &gc
}
