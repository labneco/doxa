// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type TestFieldMap FieldMap

func (f TestFieldMap) FieldMap(...FieldNamer) FieldMap {
	return FieldMap(f)
}

func TestIllegalMethods(t *testing.T) {
	Convey("Checking that invalid data leads to panic", t, func() {
		So(func() { Registry.MustGet("NonExistentModel") }, ShouldPanic)

		userModel := Registry.MustGet("User")
		So(func() { userModel.Fields().MustGet("NonExistentField") }, ShouldPanic)
		So(func() { userModel.Methods().MustGet("NonExistentMethod") }, ShouldPanic)

		So(func() { userModel.AddMethod("WrongType", "Test with int instead of func literal", 12) }, ShouldPanic)
		So(func() {
			userModel.AddMethod("ComputeAge", "Trying to add existing method", func(rc *RecordCollection) {})
		}, ShouldPanic)
		So(func() {
			userModel.AddMethod("Create", "Trying to add existing method", func(rc *RecordCollection) {})
		}, ShouldPanic)
		So(func() { userModel.AddEmptyMethod("ComputeAge") }, ShouldPanic)
		So(func() { userModel.methods.MustGet("ComputeAge").Extend("Test with int instead of func literal", 12) }, ShouldPanic)
		So(func() {
			userModel.methods.MustGet("ComputeAge").Extend("Test with wrong signature", func(rc string) (int, bool) { return 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.methods.MustGet("ComputeAge").Extend("Test with wrong signature", func(rc *RecordCollection, x string) (int, bool) { return 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.methods.MustGet("ComputeAge").Extend("Test with wrong signature", func(rc *RecordCollection) (int, int, bool) { return 0, 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.methods.MustGet("ComputeAge").Extend("Test with wrong signature", func(rc *RecordCollection) (int, bool) { return 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.methods.MustGet("DecorateEmail").Extend("Test with wrong signature", func(rc *RecordCollection, email []byte) string { return "" })
		}, ShouldPanic)
	})
	Convey("Test checkTypesMatch", t, func() {
		type TestRecordSet struct {
			*RecordCollection
		}

		var _ FieldMapper = TestFieldMap{}

		So(checkTypesMatch(reflect.TypeOf("bar"), reflect.TypeOf("bar")), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(0), reflect.TypeOf("bar")), ShouldBeFalse)
		So(checkTypesMatch(reflect.TypeOf(new(RecordCollection)), reflect.TypeOf(TestRecordSet{})), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(TestRecordSet{}), reflect.TypeOf(new(RecordCollection))), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(TestFieldMap{}), reflect.TypeOf(FieldMap{})), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(FieldMap{}), reflect.TypeOf(TestFieldMap{})), ShouldBeTrue)
	})
	Convey("Test methods signature check", t, func() {
		userModel := Registry.MustGet("User")
		nameField := userModel.Fields().MustGet("Name")
		nameField.SetOnchange(userModel.Methods().MustGet("ComputeAge"))
		processUpdates()
		So(func() { checkOnChangeMethType(nameField, "Onchange") }, ShouldPanic)
		nameField.SetOnchange(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		So(func() { checkOnChangeMethType(nameField, "Onchange") }, ShouldPanic)
		nameField.SetOnchange(userModel.Methods().MustGet("InverseSetAge"))
		processUpdates()
		So(func() { checkOnChangeMethType(nameField, "Onchange") }, ShouldPanic)
		nameField.SetOnchange(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()

		ageField := userModel.Fields().MustGet("Age")
		ageField.SetCompute(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()
		So(func() { checkComputeMethType(ageField, "Compute") }, ShouldPanic)
		ageField.SetCompute(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		So(func() { checkComputeMethType(ageField, "Compute") }, ShouldPanic)
		ageField.SetCompute(userModel.Methods().MustGet("InverseSetAge"))
		processUpdates()
		So(func() { checkComputeMethType(ageField, "Compute") }, ShouldPanic)
		ageField.SetCompute(userModel.Methods().MustGet("ComputeAge"))
		processUpdates()
	})
}

func TestBootStrap(t *testing.T) {
	// Creating a dummy table to check that it is correctly removed by Bootstrap
	dbExecuteNoTx("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")

	Convey("Database creation should run fine", t, func() {
		Convey("Dummy table should exist", func() {
			So(testAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		})
		Convey("Bootstrap should not panic", func() {
			So(BootStrap, ShouldNotPanic)
			So(SyncDatabase, ShouldNotPanic)
		})
		Convey("Boostrapping twice should panic", func() {
			So(BootStrapped(), ShouldBeTrue)
			So(BootStrap, ShouldPanic)
		})
		Convey("Creating methods after bootstrap should panic", func() {
			So(func() {
				Registry.MustGet("User").AddMethod("NewMethod", "Method after boostrap", func(rc *RecordCollection) {})
			}, ShouldPanic)
		})
		Convey("Creating SQL view should run fine", func() {
			So(func() {
				dbExecuteNoTx(`DROP VIEW IF EXISTS user_view;
					CREATE VIEW user_view AS (
						SELECT u.id, u.name, p.city, u.active
						FROM "user" u
							LEFT JOIN "profile" p ON p.id = u.profile_id
					)`)
			}, ShouldNotPanic)
		})
		Convey("All models should have a DB table", func() {
			dbTables := testAdapter.tables()
			for tableName, mi := range Registry.registryByTableName {
				if mi.isMixin() || mi.isManual() {
					continue
				}
				So(dbTables[tableName], ShouldBeTrue)
			}
		})
		Convey("All DB tables should have a model", func() {
			for dbTable := range testAdapter.tables() {
				So(Registry.registryByTableName, ShouldContainKey, dbTable)
			}
		})
		Convey("Table constraints should have been created", func() {
			So(testAdapter.constraints("%_mancon"), ShouldHaveLength, 1)
			So(testAdapter.constraints("%_mancon")[0], ShouldEqual, "nums_premium_user_mancon")
		})
		Convey("Applying DB modifications", func() {
			Registry.bootstrapped = false
			contentField := Registry.MustGet("Post").Fields().MustGet("Content")
			contentField.SetRequired(false)
			profileField := Registry.MustGet("User").Fields().MustGet("Profile")
			profileField.SetRequired(false)
			numsField := Registry.MustGet("User").Fields().MustGet("Nums")
			numsField.SetDefault(nil).SetIndex(false)
			So(BootStrap, ShouldNotPanic)
			So(contentField.required, ShouldBeFalse)
			So(profileField.required, ShouldBeFalse)
			So(numsField.index, ShouldBeFalse)
			So(SyncDatabase, ShouldNotPanic)
		})
	})

	Convey("Post testing models modifications", t, func() {
		visibilityField := Registry.MustGet("Post").Fields().MustGet("Visibility")
		So(visibilityField.selection, ShouldHaveLength, 3)
		So(visibilityField.selection, ShouldContainKey, "visible")
		So(visibilityField.selection, ShouldContainKey, "invisible")
		So(visibilityField.selection, ShouldContainKey, "logged_in")
		genderField := Registry.MustGet("Profile").Fields().MustGet("Gender")
		So(genderField.selection, ShouldHaveLength, 2)
		So(genderField.selection, ShouldContainKey, "m")
		So(genderField.selection, ShouldContainKey, "f")
	})

	Convey("Truncating all tables...", t, func() {
		for tn, mi := range Registry.registryByTableName {
			if mi.isMixin() || mi.isManual() {
				continue
			}
			dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
		}
	})
}
