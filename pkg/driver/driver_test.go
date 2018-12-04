package driver

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/LINBIT/linstor-csi/pkg/client"
	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

var controllers = flag.String("controllers", "",
	"Run suite against a real LINSTOR cluster with the specificed controller endpoints")
var node = flag.String("node", "fake.node",
	"Node ID to pass to tests, if you're running against a real LINSTOR cluster this needs to match the name of one of the real satellites")
var storagePool = flag.String("storage-pool", "", "Linstor Storage Pool for use during testing")
var endpoint = flag.String("Endpoint", "unix:///tmp/csi.sock", "Unix socket for CSI communication")
var mountForReal = flag.Bool("mount-for-real", false, "Actually try to mount volumes, needs to be ran on on a kubelet (indicted by the node flag) with it's /dev dir bind mounted into the container")

func TestDriver(t *testing.T) {

	logFile, err := ioutil.TempFile("", "csi-test-logs")
	if err != nil {
		t.Fatal(err)
	}

	driverCfg := Config{
		Endpoint: *endpoint,
		Node:     *node,
		LogOut:   logFile,
	}

	mockStorageBackend := &client.MockStorage{}
	driverCfg.Storage = mockStorageBackend
	driverCfg.Assignments = mockStorageBackend
	driverCfg.Mount = mockStorageBackend

	if *controllers != "" {
		realStorageBackend := client.NewLinstor(driverCfg.LogOut, "csi-test-annotations")
		realStorageBackend.DefaultControllers = *controllers
		realStorageBackend.DefaultStoragePool = *storagePool
		driverCfg.Storage = realStorageBackend
		driverCfg.Assignments = realStorageBackend
		if *mountForReal {
			driverCfg.Mount = realStorageBackend
		}
	}

	driver, _ := NewDriver(driverCfg)
	defer driver.Stop()

	// run your driver
	go driver.Run()

	mntDir, err := ioutil.TempDir("", "mnt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mntDir)

	mntStageDir, err := ioutil.TempDir("", "mnt-stage")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mntStageDir)

	cfg := &sanity.Config{
		StagingPath: mntStageDir,
		TargetPath:  mntDir,
		Address:     *endpoint,

		TestVolumeParameters: map[string]string{
			"autoPlace": "2",
		},
	}

	// Now call the test suite
	sanity.Test(t, cfg)
}
