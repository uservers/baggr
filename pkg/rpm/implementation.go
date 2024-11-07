package rpm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/util"

	"github.com/sirupsen/logrus"
	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/source"
	"github.com/uservers/baggr/pkg/spec"
)

const DefaultDownloadURL = "http://www.ulabs.uservers.net/no-url"

type Implementation interface {
	BuildRpmSpec(context.Context, *build.Options, source.Writer, *spec.Manifest) (string, error)
	CopySourceFiles(context.Context, *build.Options, source.Writer, *spec.Manifest) error
	BuildRpms(context.Context, *build.Options, string, source.Writer) (build.Result, error)
	VerifyPackages() error
}

type defaultImplementation struct{}

// CheckSourceFiles
func (di *defaultImplementation) CheckSourceFiles(cx context.Context, manifest *spec.Manifest) error {
	// TODO: Extract cwd from context
	notFound := []string{}
	for _, file := range manifest.Files {
		if _, err := os.Stat(file.Source); err != nil {
			if err != os.ErrNotExist {
				return fmt.Errorf("error checking for file: %w", err)
			} else {
				notFound = append(notFound, file.Source)
			}
		}
	}
	if len(notFound) > 0 {
		return fmt.Errorf("source files not found: %v", notFound)
	}
	return nil

}

// BuildRpmSpec builds the RPM spec file from the manfiest data and returns the path
func (di *defaultImplementation) BuildRpmSpec(
	ctx context.Context, opts *build.Options, sourceWriter source.Writer, omanifest *spec.Manifest,
) (string, error) {
	if len(omanifest.Files) == 0 {
		return "", fmt.Errorf("unable to build spec, no files defined in top level project")
	}

	// Get the package version
	var ver = opts.Version
	// if there is no specific version set in the options, then
	// we MUST have an autocomputed version in the options.
	if opts.Version == nil {
		buildContext := ctx.Value(build.ContextKey{})
		if buildContext == nil {
			return "", errors.New("unable to read build context")
		}

		if ver = buildContext.(build.Context).Version; ver == nil {
			return "", errors.New("no version set in options or found in build context")
		}
	}

	// Since we're altering the manifest, clone it as not to modify the original
	manifest := omanifest.DeepCopy()
	if manifest.URL == "" {
		manifest.URL = DefaultDownloadURL
	}

	// requires can be:   perl >= 9:5.00502-3
	// http://ftp.rpm.org/api/4.4.2.2/dependencies.html

	// Build a %prep stage that copies all files. This cannot be done outside
	// of the spec because they get deleted at %install time

	// List of directories to be created empty
	buildrootDirectoryList := map[string]string{}

	// prep_file_commands is the list of commands that will be executed in
	// the %prep section of the RPM build
	prep_file_commands := ""

	// When building, rpmbuild will cd to the BUILD directory. This means that the repo
	// directory is lost in the ether. Hence, we need to capture this as soon as possible,
	// possibly during __init__.

	// Normalize the file records
	manifest.EnsureDefaults()

	// Process the main component files
	p2, b2 := processComponentFiles(sourceWriter, &manifest.Component)
	prep_file_commands += p2
	for _, p := range b2 {
		buildrootDirectoryList[p] = p
	}

	for i := range manifest.Components {
		p2, b2 := processComponentFiles(sourceWriter, manifest.Components[i])
		prep_file_commands += p2
		for _, p := range b2 {
			buildrootDirectoryList[p] = p
		}
	}

	for dirname := range buildrootDirectoryList {
		prep_file_commands += fmt.Sprintf("%%{__mkdir_p} %%{buildroot}%s || exit 111\n", dirname)
	}

	tmpl, err := template.New("template.tmpl").Parse(Template)

	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	f, err := os.CreateTemp("", "rpmbuilder-*.spec")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}

	if err := tmpl.Execute(f, map[string]interface{}{
		"Manifest":           manifest,
		"PrepFileCommands":   prep_file_commands,
		"Version":            ver.String,
		"Release":            ver.Release,
		"BuildrootDirectory": buildrootDirectoryList,
	}); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	logrus.Infof("Wrote RPM spec file to %s", f.Name())
	return f.Name(), nil
}

// processComponentFiles
func processComponentFiles(sourceWriter source.Writer, component *spec.Component) (string, map[string]string) {
	buildrootDirectoryList := map[string]string{}
	prepFileCommands := ""
	for _, filedata := range component.Files {
		// Handle directories, either empty/non-existent (expressed with '%DIR%') or
		// already existing in the sourceWriter path, in which case they will be
		// copied recursively.
		if filedata.Source == "%DIR%" || util.IsDir(filepath.Join(sourceWriter.Path(), filedata.Destination)) {
			// if filedata.Destination == "" {
			// 	logrus.Warnf("skipping file #%d from component %q as it is empty dir with no destination", i, component.Name)
			// 	continue
			// }
			prepFileCommands += fmt.Sprintf("%%{__mkdir_p} %%{buildroot}/%s\n", filedata.Destination)

			// If the file entry source exists in the sourcewriter's path, add
			// instructions to the spec to copy it recursively:
			if filedata.Source != "%DIR%" && util.IsDir(filepath.Join(sourceWriter.Path(), filedata.Destination)) {
				prepFileCommands += fmt.Sprintf(
					"%%{__cp} -rL %s/* $RPM_BUILD_ROOT/%s || :\n",
					filepath.Join(sourceWriter.Path(), filedata.Source),
					filedata.Destination,
				)
			}

			if filedata.Source != "%DIR%" {
				dirpath := path.Clean(filedata.Destination)
				buildrootDirectoryList[dirpath] = dirpath
			}
			continue
		}

		// Handle files.
		// At this point, the sourceWriter has already copied the files from
		// the source reader, so they are in the right location
		// realPath := filepath.Join(sourceWriter.Path(), filedata.Destination)
		// if filedata.Destination == "" {
		// 	realPath = filepath.Join(sourceWriter.Path(), filedata.Source)
		// }
		destPath := filedata.Destination
		if destPath == "" {
			destPath = filedata.Source
		}

		// Antes el RPM copiaba, pero ahora lo hace el source,,, m,mhhh
		// prepFileCommands += fmt.Sprintf("%%{__cp} -p -L ")
		// prepFileCommands += fmt.Sprintf("%s ", realPath)
		// prepFileCommands += fmt.Sprintf("%%{buildroot}%s || exit 111\n", destPath)

		// Register the path in the full list
		dirpath := path.Clean(destPath)
		dirname := path.Dir(dirpath)
		buildrootDirectoryList[dirname] = dirname
	}
	return prepFileCommands, buildrootDirectoryList
}

// BuildRpms builds the RPMs packages shelling out to rpmbuild
func (di *defaultImplementation) BuildRpms(
	ctx context.Context, opts *build.Options, specPath string, sourceWriter source.Writer,
) (results build.Result, err error) {
	// tmp, err := os.CreateTemp("", "broo-*")
	//  Este es el comando que vamos a correr:
	rpmProc := command.New(
		"rpmbuild", "-bb", specPath, "-vv", "--buildroot", sourceWriter.Path(), "--target", "noarch", // 2>&1",
	)

	// Execute rpmbuild
	output, err := rpmProc.RunSilentSuccessOutput()
	if err != nil {
		return results, fmt.Errorf("executing rpmbuild: %w", err)
	}

	// Build the results set
	results = build.Result{
		Artifacts: []build.Artifact{},
		Log:       output.OutputTrimNL(),
		Error:     errors.New(output.Error()),
	}

	// Scan output to find built rpms
	for _, l := range findFiles(output.OutputTrimNL()) {
		results.Artifacts = append(results.Artifacts, build.NewFileArtifact(l))
	}

	return results, nil
}

// findFiles parses the build output to find the built files
// TODO(puerco): Maybe create a temp dir and get all files from there.
func findFiles(buildLog string) []string {
	ret := []string{}
	lines := strings.Split(buildLog, "\n")
	r := regexp.MustCompile(`^Wrote:\s(\S+\.rpm)$`)
	for _, l := range lines {
		if strs := r.FindStringSubmatch(l); len(strs) == 2 {
			ret = append(ret, strs[1])
		}
	}
	return ret
}

// CopySourceFilescopies the files from the source reader using the source writer
func (di *defaultImplementation) CopySourceFiles(ctx context.Context, opts *build.Options, sourceWriter source.Writer, manifest *spec.Manifest) error {
	if opts.SourceReader == nil {
		return fmt.Errorf("unable to copy files, no source reader defined")
	}

	// Copy the main component paths
	if err := sourceWriter.CopyPaths(ctx, opts.SourceReader, manifest.Component.Files); err != nil {
		return fmt.Errorf("copying main component files: %w", err)
	}

	for _, c := range manifest.Components {
		if err := sourceWriter.CopyPaths(ctx, opts.SourceReader, c.Files); err != nil {
			return fmt.Errorf("copying files from %q: %w", c.Name, err)
		}
	}

	return nil
}

func (di *defaultImplementation) VerifyPackages() error { return nil }
