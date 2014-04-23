package urknall

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/cmd"
)

func (rl *Package) installBinaryPackage() {
	bpkg := rl.pkg.(BinaryPackage)
	// Download package and checksum.
	repo := rl.host.BinaryPackageRepository

	pkgPath := fmt.Sprintf("/tmp/%s.%s.$(uname -i).bpkg", bpkg.Name(), bpkg.PkgVersion())

	rl.Add(
		fmt.Sprintf(`curl -SsfL %s/get/%s/%s/$(uname -i)/data -o %s || { echo "package does not exist"; false; }`, repo, bpkg.Name(), bpkg.PkgVersion(), pkgPath),
		// Validate downloaded archive.
		fmt.Sprintf(`echo "$(curl -SsfL %s/get/%s/%s/$(uname -i)/checksum)  %s" | sha256sum -c -`, repo, bpkg.Name(), bpkg.PkgVersion(), pkgPath),
		// Validate package architecture.
		fmt.Sprintf(`[[ $(ar p %s bpkg.metadata | grep Architecture | cut -d" " -f2) == $(uname -i) ]]`, pkgPath),
		// Extract data.tar.gz and verify checksum.
		fmt.Sprintf(`ar p %s data.tar.gz > /tmp/data.tar.gz`, pkgPath),
		fmt.Sprintf(`echo "$(ar p %s bpkg.metadata | grep Checksum | cut -d" " -f2)  /tmp/data.tar.gz" | sha256sum -c -`, pkgPath),

		// Install package dependencies and binary package.
		fmt.Sprintf(`DEBIAN_FRONTEND=noninteractive apt-get install -y $(ar -p %s bpkg.metadata | grep Depends | cut -d" " -f2-)`, pkgPath),
		"gunzip -c /tmp/data.tar.gz | tar -C / -x",
		// Clean up /tmp folder.
		fmt.Sprintf(`rm -f /tmp/data.tar.gz %s`, pkgPath),
	)
}

func (rl *Package) buildBinaryPackage() (e error) {
	repo := rl.host.BinaryPackageRepository

	bpkg := rl.pkg.(BinaryPackage)
	installPath := bpkg.InstallPath()

	if !strings.HasPrefix(installPath, "/opt") {
		return fmt.Errorf("currently installation is only allowed to the /opt directory")
	}

	tmpDir := "/tmp/" + rl.name

	metadata := []string{}
	metadata = append(metadata, "Package: "+bpkg.Name())
	metadata = append(metadata, "Version: "+bpkg.PkgVersion())
	metadata = append(metadata, "Depends: "+strings.Join(bpkg.PackageDependencies(), " "))
	metadata = append(metadata, "Checksum: CHECKSUM")
	metadata = append(metadata, "Architecture: ARCH")

	pkgFilename := fmt.Sprintf("%s.%s.$(uname -i).bpkg", bpkg.Name(), bpkg.PkgVersion())

	rl.Add(
		fmt.Sprintf(`curl -SsfL %s/get/%s/%s/$(uname -i)/checksum > /dev/null && { echo "package already exists"; false; }`, repo, bpkg.Name(), bpkg.PkgVersion()),
		// Create tmp file folder.
		cmd.Mkdir(tmpDir, "root", 0755),
		// Create archive.
		"tar cfz "+tmpDir+"/data.tar.gz "+installPath,
		// Create and fill metadata file.
		cmd.WriteFile(tmpDir+"/bpkg.metadata", strings.Join(metadata, "\n"), "root", 0644),
		fmt.Sprintf(`sed -e "s/CHECKSUM/$(sha256sum -b %[1]s/data.tar.gz | cut -d' ' -f1)/" -i %[1]s/bpkg.metadata`, tmpDir),
		fmt.Sprintf(`sed -e "s/ARCH/$(uname -i)/" -i %s/bpkg.metadata`, tmpDir),
		// Create archive.
		fmt.Sprintf("ar cr %[1]s/%[2]s %[1]s/bpkg.metadata %[1]s/data.tar.gz", tmpDir, pkgFilename),
		// Push archive to repository.
		fmt.Sprintf("curl -Ssf -F data=@%[1]s/%[2]s -F file=%[2]s %[3]s/add", tmpDir, pkgFilename, repo),
		// Clean up the temp files.
		fmt.Sprintf("rm -f %[1]s/bpkg.metdata %[1]s/data.tar.gz %[1]s/%[2]s", tmpDir, pkgFilename),
		"rm -Rf "+tmpDir,
	)

	return nil
}
