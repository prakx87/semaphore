package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db_lib"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

func (t *LocalJob) installInventory() (err error) {
	if t.Inventory.SSHKeyID != nil {
		t.sshKeyInstallation, err = t.Inventory.SSHKey.Install(db.AccessKeyRoleAnsibleUser, t.Logger)
		if err != nil {
			return
		}
	}

	if t.Inventory.BecomeKeyID != nil {
		t.becomeKeyInstallation, err = t.Inventory.BecomeKey.Install(db.AccessKeyRoleAnsibleBecomeUser, t.Logger)
		if err != nil {
			return
		}
	}

	if t.Inventory.Type == db.InventoryFile {
		err = t.cloneInventoryRepo()
	} else if t.Inventory.Type == db.InventoryStatic || t.Inventory.Type == db.InventoryStaticYaml {
		err = t.installStaticInventory()
	}

	return
}

func (t *LocalJob) tmpInventoryFilename() string {
	return "inventory_" + strconv.Itoa(t.Task.ID)
}

func (t *LocalJob) tmpInventoryFullPath() string {
	pathname := path.Join(util.Config.TmpPath, t.tmpInventoryFilename())
	if t.Inventory.Type == db.InventoryStaticYaml {
		pathname += ".yml"
	}
	return pathname
}

func (t *LocalJob) cloneInventoryRepo() error {
	if t.Inventory.Repository == nil {
		return nil
	}

	t.Log("cloning inventory repository")

	repo := db_lib.GitRepository{
		Logger:     t.Logger,
		TmpDirName: t.tmpInventoryFilename(),
		Repository: *t.Inventory.Repository,
		Client:     db_lib.CreateDefaultGitClient(),
	}

	err := repo.Clone()

	return err
}

func (t *LocalJob) installStaticInventory() error {
	t.Log("installing static inventory")

	path := t.tmpInventoryFullPath()

	// create inventory file
	return os.WriteFile(path, []byte(t.Inventory.Inventory), 0664)
}

func (t *LocalJob) destroyInventoryFile() {
	path := t.tmpInventoryFullPath()
	if err := os.Remove(path); err != nil {
		log.Error(err)
	}
}

func (t *LocalJob) destroyKeys() {
	err := t.sshKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory user key, error: " + err.Error())
	}

	err = t.becomeKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory become user key, error: " + err.Error())
	}

	err = t.vaultFileInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory vault password file, error: " + err.Error())
	}
}
