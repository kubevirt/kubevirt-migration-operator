/*
Copyright 2025 The KubeVirt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	operator "kubevirt.io/kubevirt-migration-operator/pkg/resources/operator"
)

var (
	csvVersion         = flag.String("csv-version", "", "")
	replacesCsvVersion = flag.String("replaces-csv-version", "", "")
	namespace          = flag.String("namespace", "", "")
	pullPolicy         = flag.String("pull-policy", "", "")

	logoBase64 = flag.String("logo-base64", "", "")
	verbosity  = flag.String("verbosity", "1", "")

	operatorVersion = flag.String("operator-version", "", "")

	operatorImage   = flag.String("operator-image", "", "")
	controllerImage = flag.String("controller-image", "", "")
	dumpCRDs        = flag.Bool("dump-crds", false, "optional - dumps migration operator related crd manifests to stdout")
)

func main() {
	flag.Parse()

	data := operator.ClusterServiceVersionData{
		CsvVersion:         *csvVersion,
		ReplacesCsvVersion: *replacesCsvVersion,
		Namespace:          *namespace,
		ImagePullPolicy:    *pullPolicy,
		IconBase64:         *logoBase64,
		Verbosity:          *verbosity,

		OperatorVersion: *operatorVersion,

		ControllerImage: *controllerImage,
		OperatorImage:   *operatorImage,
	}

	csv, err := operator.NewClusterServiceVersion(&data)
	if err != nil {
		panic(err)
	}
	if err = marshallObject(csv, os.Stdout); err != nil {
		panic(err)
	}

	if *dumpCRDs {
		migCtrlCRD := operator.NewMigControllerCrd()
		if err = marshallObject(migCtrlCRD, os.Stdout); err != nil {
			panic(err)
		}
	}
}
