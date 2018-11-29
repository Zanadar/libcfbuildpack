/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buildpack_test

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestLicense(t *testing.T) {
	spec.Run(t, "License", testLicense, spec.Random(), spec.Report(report.Terminal{}))
}

func testLicense(t *testing.T, when spec.G, it spec.S) {

	it("validates with type set", func() {
		err := buildpack.License{Type: "test-type"}.Validate()
		if err != nil {
			t.Errorf("License.Validate() = %s expected no error", err)
		}
	})

	it("validates with uri set", func() {
		err := buildpack.License{URI: "test-uri"}.Validate()
		if err != nil {
			t.Errorf("License.Validate() = %s expected no error", err)
		}
	})

	it("validates with type and uri set", func() {
		err := buildpack.License{Type: "test-type", URI: "test-uri"}.Validate()
		if err != nil {
			t.Errorf("License.Validate() = %s expected no error", err)
		}
	})

	it("does not validate without type and uri set", func() {
		err := buildpack.License{}.Validate()
		if err == nil {
			t.Errorf("License.Validate() = nil expected error")
		}
	})
}
