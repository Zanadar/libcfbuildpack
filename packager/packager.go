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

package packager

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	buildpackBp "github.com/buildpack/libbuildpack/buildpack"
	layersBp "github.com/buildpack/libbuildpack/layers"
	loggerBp "github.com/buildpack/libbuildpack/logger"
	buildpackCf "github.com/cloudfoundry/libcfbuildpack/buildpack"
	layersCf "github.com/cloudfoundry/libcfbuildpack/layers"
	loggerCf "github.com/cloudfoundry/libcfbuildpack/logger"
)

// Packager is a root element for packaging up a buildpack
type Packager struct {
	Buildpack buildpackCf.Buildpack
	Layers    layersCf.Layers
	Logger    loggerCf.Logger
}

// Create creates a new buildpack package.
func (p Packager) Create() error {
	p.Logger.FirstLine("Packaging %s", p.Logger.PrettyVersion(p.Buildpack))

	if err := p.prePackage(); err != nil {
		return err
	}

	includedFiles, err := p.Buildpack.IncludeFiles()
	if err != nil {
		return err
	}

	dependencyFiles, err := p.cacheDependencies()
	if err != nil {
		return err
	}

	return p.createArchive(append(includedFiles, dependencyFiles...))
}

func (p Packager) addFile(out *tar.Writer, path string) error {
	p.Logger.SubsequentLine("Adding %s", path)

	file, err := os.Open(filepath.Join(p.Buildpack.Root, path))
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := new(tar.Header)
	header.Name = path
	header.Size = stat.Size()
	header.Mode = int64(stat.Mode())
	header.ModTime = stat.ModTime()

	if err := out.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(out, file)
	return err
}

func (p Packager) archivePath() (string, error) {
	dir, err := osArgs(1)
	if err != nil {
		return "", err
	}

	info := p.Buildpack.Info

	path := []string{dir}
	path = append(path, strings.Split(info.ID, ".")...)
	path = append(path, info.ID, info.Version)

	f := fmt.Sprintf("%s-%s.tgz", info.ID, info.Version)
	f = strings.Replace(f, "SNAPSHOT", fmt.Sprintf("%s-1", time.Now().Format("20060102.150405")), 1)

	path = append(path, f)

	return filepath.Join(path...), nil
}

func (p Packager) cacheDependencies() ([]string, error) {
	var files []string

	deps, err := p.Buildpack.Dependencies()
	if err != nil {
		return nil, err
	}

	for _, dep := range deps {
		p.Logger.FirstLine("Caching %s", p.Logger.PrettyVersion(dep))

		layer := p.Layers.DownloadLayer(dep)

		a, err := layer.Artifact()
		if err != nil {
			return nil, err
		}

		artifact, err := filepath.Rel(p.Buildpack.Root, a)
		if err != nil {
			return nil, err
		}

		metadata, err := filepath.Rel(p.Buildpack.Root, layer.Metadata)
		if err != nil {
			return nil, err
		}

		files = append(files, artifact, metadata)
	}

	return files, nil
}

func (p Packager) createArchive(files []string) error {
	archive, err := p.archivePath()
	if err != nil {
		return err
	}

	p.Logger.FirstLine("Creating archive %s", archive)

	if err = os.MkdirAll(filepath.Dir(archive), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(archive, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		if err := p.addFile(tw, file); err != nil {
			return err
		}
	}

	return nil
}

func (p Packager) prePackage() error {
	pp, ok := p.Buildpack.PrePackage()
	if !ok {
		return nil
	}

	cmd := exec.Command(pp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = p.Buildpack.Root

	p.Logger.FirstLine("Pre-Package with %s", strings.Join(cmd.Args, " "))

	return cmd.Run()
}

// DefaultPackager creates a new Packager, using the executable to find the root of the buildpack.
func DefaultPackager() (Packager, error) {
	l := loggerBp.DefaultLogger()
	logger := loggerCf.Logger{Logger: l}

	b, err := buildpackBp.DefaultBuildpack(l)
	if err != nil {
		return Packager{}, err
	}
	buildpack := buildpackCf.NewBuildpack(b)

	layers := layersCf.Layers{Layers: layersBp.Layers{Root: buildpack.CacheRoot}}

	return Packager{buildpack, layers, logger}, nil
}

func osArgs(index int) (string, error) {
	if len(os.Args) < index+1 {
		return "", fmt.Errorf("incorrect number of command line arguments")
	}

	return os.Args[index], nil
}
