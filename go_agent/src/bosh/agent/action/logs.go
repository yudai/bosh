package action

import (
	boshblob "bosh/blobstore"
	bosherr "bosh/errors"
	boshcmd "bosh/platform/commands"
	boshsettings "bosh/settings"
	"path/filepath"
)

type logsAction struct {
	compressor boshcmd.Compressor
	blobstore  boshblob.Blobstore
}

func newLogs(compressor boshcmd.Compressor, blobstore boshblob.Blobstore) (action logsAction) {
	action.compressor = compressor
	action.blobstore = blobstore
	return
}

func (a logsAction) IsAsynchronous() bool {
	return true
}

func (a logsAction) Run(logType string, filters []string) (value interface{}, err error) {
	var logsDir string

	switch logType {
	case "job":
		if len(filters) == 0 {
			filters = []string{"**/*.log"}
		}
		logsDir = filepath.Join(boshsettings.VCAP_BASE_DIR, "sys", "log")
	case "agent":
		if len(filters) == 0 {
			filters = []string{"**/*"}
		}
		logsDir = filepath.Join(boshsettings.VCAP_BASE_DIR, "bosh", "log")
	default:
		err = bosherr.New("Invalid log type")
		return
	}

	tarball, err := a.compressor.CompressFilesInDir(logsDir, filters)
	if err != nil {
		err = bosherr.WrapError(err, "Making logs tarball")
		return
	}

	blobId, _, err := a.blobstore.Create(tarball)
	if err != nil {
		err = bosherr.WrapError(err, "Create file on blobstore")
		return
	}

	value = map[string]string{"blobstore_id": blobId}
	return
}
