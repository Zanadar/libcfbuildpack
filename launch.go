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

package libjavabuildpack

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/buildpack/libbuildpack"
	"github.com/fatih/color"
)

// Launch is an extension to libbuildpack.Launch that allows additional functionality to be added.
type Launch struct {
	libbuildpack.Launch

	// Cache is the Cache to use to acquire dependencies.
	Cache Cache

	// Logger logger is used to write debug and info to the console.
	Logger Logger
}

// DependencyLayer returns a DependencyLaunchLayer unique to a dependency.
func (l Launch) DependencyLayer(dependency Dependency) DependencyLaunchLayer {
	return DependencyLaunchLayer{
		l.Layer(dependency.ID),
		l.Logger,
		dependency,
		l.Cache.DownloadLayer(dependency),
	}
}

// String makes Launch satisfy the Stringer interface.
func (l Launch) String() string {
	return fmt.Sprintf("Launch{ Launch: %s Cache: %s, Logger: %s }", l.Launch, l.Cache, l.Logger)
}

// WriteMetadata writes Launch metadata to the filesystem.
func (l Launch) WriteMetadata(metadata libbuildpack.LaunchMetadata) error {
	l.Logger.FirstLine("Process types:")

	max := l.maximumTypeLength(metadata)

	for _, t := range metadata.Processes {
		s := color.CyanString(t.Type) + ":"

		for i := 0; i < (max - len(t.Type)); i++ {
			s += " "
		}

		l.Logger.SubsequentLine("%s %s", s, t.Command)
	}

	return l.Launch.WriteMetadata(metadata)
}

func (l Launch) maximumTypeLength(metadata libbuildpack.LaunchMetadata) int {
	max := 0

	for _, t := range metadata.Processes {
		l := len(t.Type)

		if l > max {
			max = l
		}
	}

	return max
}

// DependencyLaunchLayer is an extension to LaunchLayer that is unique to a dependency.
type DependencyLaunchLayer struct {
	libbuildpack.LaunchLayer

	// Logger is used to write debug and info to the console.
	Logger Logger

	dependency    Dependency
	downloadLayer DownloadCacheLayer
}

// ArtifactName returns the name portion of the download path for the dependency.
func (d DependencyLaunchLayer) ArtifactName() string {
	return filepath.Base(d.dependency.URI)
}

// String makes DependencyLaunchLayer satisfy the Stringer interface.
func (d DependencyLaunchLayer) String() string {
	return fmt.Sprintf("DependencyLaunchLayer{ LaunchLayer: %s, Logger: %s, dependency: %s, downloadLayer: %s }",
		d.LaunchLayer, d.Logger, d.dependency, d.dependency)
}

// LaunchContributor defines a callback function that is called when a dependency needs to be contributed.
type LaunchContributor func(artifact string, layer DependencyLaunchLayer) error

// Contribute facilitates custom contribution of an artifact to a launch layer.  If the artifact has already been
// contributed, the contribution is validated and the contributor is not called.
func (d DependencyLaunchLayer) Contribute(contributor LaunchContributor) error {
	var m Dependency

	if err := d.ReadMetadata(&m); err != nil {
		d.Logger.Debug("Dependency metadata is not structured correctly")
		return err
	}

	if reflect.DeepEqual(d.dependency, m) {
		d.Logger.FirstLine("%s: %s cached launch layer",
			d.Logger.PrettyVersion(d.dependency), color.GreenString("Reusing"))
		return nil
	}

	d.Logger.Debug("Download metadata %s does not match expected %s", m, d.dependency)

	d.Logger.FirstLine("%s: %s to launch",
		d.Logger.PrettyVersion(d.dependency), color.YellowString("Contributing"))

	if err := os.RemoveAll(d.Root); err != nil {
		return err
	}

	a, err := d.downloadLayer.Artifact()
	if err != nil {
		return err
	}

	if err := contributor(a, d); err != nil {
		d.Logger.Debug("Error during contribution")
		return err;
	}

	return d.WriteMetadata(d.dependency)
}

// WriteProfile writes a file to profile.d with this value.
func (d DependencyLaunchLayer) WriteProfile(file string, format string, args ...interface{}) error {
	d.Logger.SubsequentLine("Writing .profile.d/%s", file)
	return d.LaunchLayer.WriteProfile(file, format, args...)
}
