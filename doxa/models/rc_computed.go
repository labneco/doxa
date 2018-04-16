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

package models

import (
	"github.com/labneco/doxa/doxa/models/security"
	"github.com/labneco/doxa/doxa/tools/typesutils"
)

// computeFieldValues updates the given params with the given computed (non stored) fields
// or all the computed fields of the model if not given.
// Returned fieldMap keys are field's JSON name
//
// This method reads result from cache if available. If not, the computation is carried
// out and the result is stored in cache.
func (rc *RecordCollection) computeFieldValues(params *FieldMap, fields ...string) {
	rc.EnsureOne()
	for _, fInfo := range rc.model.fields.getComputedFields(fields...) {
		if !checkFieldPermission(fInfo, rc.env.uid, security.Read) {
			// We do not have the access rights on this field, so we skip it.
			continue
		}
		if _, exists := (*params)[fInfo.name]; exists {
			// We already have the value we need in params
			// probably because it was computed with another field
			continue
		}
		if rc.env.cache.checkIfInCache(rc.model, rc.Ids(), []string{fInfo.name}) {
			(*params)[fInfo.json] = rc.env.cache.get(rc.model, rc.Ids()[0], fInfo.name)
			continue
		}
		newParams := rc.Call(fInfo.compute).(FieldMapper).FieldMap()
		for k, v := range newParams {
			key, _ := rc.model.fields.Get(k)
			(*params)[key.json] = v
			rc.env.cache.updateEntry(rc.model, rc.Ids()[0], fInfo.name, v)
		}
	}
}

// processTriggers execute computed fields recomputation (for stored fields) or
// invalidation (for non stored fields) based on the data of each fields 'Depends'
// attribute.
func (rc *RecordCollection) processTriggers(fMap FieldMap) {
	if rc.Env().Context().GetBool("doxa_no_recompute_stored_fields") {
		return
	}
	// Find record fields to update from the modified fields of fMap
	fieldNames := fMap.Keys()
	toUpdate := make(map[computeData][]FieldNamer)
	for _, fieldName := range fieldNames {
		refFieldInfo, ok := rc.model.fields.Get(fieldName)
		if !ok {
			continue
		}
		for _, dep := range refFieldInfo.dependencies {
			fName := dep.fieldName
			if dep.stored {
				// We remove fieldName to group calls to the compute method if stored
				dep.fieldName = ""
			}
			toUpdate[dep] = append(toUpdate[dep], FieldName(fName))
		}
	}

	// Compute all that must be computed and store the values
	rc.Fetch()
	for cData, fNames := range toUpdate {
		recs := rc
		if cData.path != "" {
			recs = rc.Env().Pool(cData.model.name).Search(rc.Model().Field(cData.path).In(rc.Ids()))
		}
		if !cData.stored {
			// Field is not stored, just invalidating cache
			for _, id := range recs.Ids() {
				rc.env.cache.removeEntry(recs.model, id, cData.fieldName)
			}
			continue
		}
		updateStoredFields(recs, cData.compute, fNames)
	}
}

// updateStoredFields calls the given computeMethod on recs and stores the values.
func updateStoredFields(recs *RecordCollection, computeMethod string, fieldsToReset []FieldNamer) {
	for _, rec := range recs.Records() {
		retVal := rec.Call(computeMethod)
		vals := retVal.(FieldMapper).FieldMap(fieldsToReset...)
		// Check if the values actually changed
		var doUpdate bool
		for f, v := range vals {
			if f == "write_date" {
				continue
			}
			if rs, isRS := rec.Get(f).(RecordSet); isRS {
				if !rs.Collection().Equals(v.(RecordSet).Collection()) {
					doUpdate = true
					break
				}
				continue
			}
			if rec.Get(f) != v {
				doUpdate = true
				break
			}
		}
		if doUpdate {
			rec.WithContext("doxa_force_compute_write", true).Call("Write", vals, fieldsToReset)
		}
	}
}

// processInverseMethods executes inverse methods of fields in the given
// FieldMap if it exists. It returns a new FieldMap to be used by Create/Write
// instead of the original one.
func (rc *RecordCollection) processInverseMethods(fMap FieldMap) {
	for fieldName := range fMap {
		fi := rc.model.getRelatedFieldInfo(fieldName)
		if !fi.isComputedField() || rc.Env().Context().HasKey("doxa_force_compute_write") {
			continue
		}
		val, exists := fMap.Get(fi.json, fi.model)
		if !exists {
			continue
		}
		if fi.inverse == "" {
			if rc.Env().Context().GetBool("doxa_allow_without_inverse") {
				continue
			}
			if typesutils.IsZero(val) {
				continue
			}
			log.Panic("Trying to write a computed field without inverse method", "model", rc.model.name, "field", fieldName)
		}
		rc.CallMulti(fi.inverse, val)
	}
}
