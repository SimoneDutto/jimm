// Copyright 2026 Canonical.

package testing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/juju/juju/api/client/keymanager"
	"github.com/juju/utils/v4/ssh"
	"github.com/juju/version/v2"
	gossh "golang.org/x/crypto/ssh"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/testutils/jimmtest"
)

// TestModelSSHKeys verifies that Juju 3 model key management is delegated to
// the backing controller, while Juju 4 model key management is persisted in
// JIMM's model-scoped SSH key database.
func TestModelSSHKeys(t *testing.T) {
	c := qt.New(t)
	s := jimmtest.SetupJimmWithControllers(c)
	model := s.CreateModelForBob(c)
	controllerVersion := version.MustParse(model.Controller.AgentVersion)
	keysAreStoredByJIMM := controllerVersion.Major >= 4

	modelTag := model.ResourceTag()
	conn := s.Open(c, nil, bobOwnerTag.Id(), &modelTag)
	defer conn.Close()
	client := keymanager.NewClient(conn)
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	c.Assert(err, qt.IsNil)
	publicKey, err := gossh.NewPublicKey(privateKey.Public())
	c.Assert(err, qt.IsNil)
	authorizedKey := fmt.Sprintf("%s %s %s", publicKey.Type(), base64.StdEncoding.EncodeToString(publicKey.Marshal()), "model-ssh-key")
	fingerprint := gossh.FingerprintLegacyMD5(publicKey)

	addResults, err := client.AddKeys(c.Context(), "unused", authorizedKey)
	c.Assert(err, qt.IsNil)
	c.Check(addResults, qt.HasLen, 0)
	keys, err := client.ListKeys(c.Context(), ssh.Fingerprints)
	c.Assert(err, qt.IsNil)
	c.Assert(keys, qt.HasLen, 1)
	c.Check(keys[0].Error, qt.IsNil)
	c.Check(keys[0].Result, qt.DeepEquals, []string{fmt.Sprintf("%s (%s)", fingerprint, "model-ssh-key")})
	if keysAreStoredByJIMM {
		dbKeys, err := s.JIMM.Database.ListSSHKeysForUser(c.Context(), bobOwnerTag.Id(), db.SSHKeyModelFilter{ModelUUID: model.UUID.String})
		c.Assert(err, qt.IsNil)
		c.Check(dbKeys, qt.HasLen, 1)
	}

	deleteResults, err := client.DeleteKeys(c.Context(), "unused", fingerprint)
	c.Assert(err, qt.IsNil)
	c.Check(deleteResults, qt.HasLen, 0)
	if keysAreStoredByJIMM {
		dbKeys, err := s.JIMM.Database.ListSSHKeysForUser(c.Context(), bobOwnerTag.Id(), db.SSHKeyModelFilter{ModelUUID: model.UUID.String})
		c.Assert(err, qt.IsNil)
		c.Check(dbKeys, qt.HasLen, 0)
	}
}
