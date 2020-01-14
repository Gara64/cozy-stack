package migrations

import (
	"fmt"
	"os"
	"testing"

	"github.com/cozy/cozy-stack/model/account"
	"github.com/cozy/cozy-stack/model/bitwarden"
	"github.com/cozy/cozy-stack/model/bitwarden/settings"
	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/model/instance/lifecycle"
	"github.com/cozy/cozy-stack/model/stack"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/tests/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMigrateSingleAccount(t *testing.T) {
	domain := "cozy.example.net"
	err := lifecycle.Destroy(domain)
	if err != instance.ErrNotFound {
		assert.NoError(t, err)
	}
	inst, err := lifecycle.Create(&lifecycle.Options{
		Domain:     "cozy.example.net",
		Passphrase: "cozy",
		PublicName: "Pierre",
	})
	assert.NoError(t, err)
	encCreds, err := account.EncryptCredentials("pierre", "topsecret")
	assert.NoError(t, err)
	basicInfo := &account.BasicInfo{
		EncryptedCredentials: encCreds,
	}
	acc := &account.Account{
		Basic: basicInfo,
	}
	setting, err := settings.Get(inst)
	assert.NoError(t, err)
	orgKey, err := setting.OrganizationKey()
	assert.NoError(t, err)

	err = couchdb.CreateDoc(inst, acc)
	assert.NoError(t, err)
	slug := "dummy"
	url := "dummy.org"

	err = migrateSingleAccount(inst, orgKey, acc, slug, url)
	assert.NoError(t, err)

	relation := acc.Relationships[consts.BitwardenCipherRelationship]
	cipherRel, ok := relation.(map[string][]VaultReference)
	assert.True(t, ok)
	cipherID := cipherRel["data"][0].ID
	assert.NotEmpty(t, cipherID)

	cipher := &bitwarden.Cipher{}
	err = couchdb.GetDoc(inst, consts.BitwardenCiphers, cipherID, cipher)
	assert.NoError(t, err)
	assert.True(t, cipher.SharedWithCozy)
}

func TestMain(m *testing.M) {
	config.UseTestFile()
	testutils.NeedCouchdb()
	if _, err := stack.Start(); err != nil {
		fmt.Printf("Error while starting the job system: %s\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
